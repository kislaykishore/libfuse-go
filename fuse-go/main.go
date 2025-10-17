package main

/*
#cgo CFLAGS: -DFUSE_USE_VERSION=29
#cgo pkg-config: fuse
#include <fuse_lowlevel.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

// Forward declarations for Go functions
void go_ll_getattr(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void go_ll_readdir(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off, struct fuse_file_info *fi);
void go_ll_lookup(fuse_req_t req, fuse_ino_t parent, char *name);
void go_ll_open(fuse_req_t req, fuse_ino_t ino, struct fuse_file_info *fi);
void go_ll_read(fuse_req_t req, fuse_ino_t ino, size_t size, off_t off, struct fuse_file_info *fi);

// C wrapper for the lookup function to handle const correctness
static void lookup_wrapper(fuse_req_t req, fuse_ino_t parent, const char *name) {
    go_ll_lookup(req, parent, (char *)name);
}

static struct fuse_lowlevel_ops ll_ops;

void init_ops() {
	ll_ops.getattr = go_ll_getattr;
	ll_ops.lookup = lookup_wrapper;
	ll_ops.readdir = go_ll_readdir;
	ll_ops.open = go_ll_open;
	ll_ops.read = go_ll_read;
}

*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

const (
	rootInode  = C.FUSE_ROOT_ID
	helloInode = 2
)

var (
	helloContent = []byte("Hello, World!\n")
	helloName    = "hello"
)

//export go_ll_getattr
func go_ll_getattr(req C.fuse_req_t, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) {
	var stat C.struct_stat
	stat.st_ino = ino
	if ino == rootInode {
		stat.st_mode = C.S_IFDIR | 0755
		stat.st_nlink = 2
	} else if ino == helloInode {
		stat.st_mode = C.S_IFREG | 0444
		stat.st_nlink = 1
		stat.st_size = C.off_t(len(helloContent))
	} else {
		C.fuse_reply_err(req, C.ENOENT)
		return
	}
	C.fuse_reply_attr(req, &stat, 1.0)
}

//export go_ll_lookup
func go_ll_lookup(req C.fuse_req_t, parent C.fuse_ino_t, name *C.char) {
	if parent != rootInode {
		C.fuse_reply_err(req, C.ENOENT)
		return
	}
	if C.GoString(name) == helloName {
		var entry C.struct_fuse_entry_param
		entry.ino = helloInode
		entry.attr_timeout = 1.0
		entry.entry_timeout = 1.0

		var stat C.struct_stat
		stat.st_ino = helloInode
		stat.st_mode = C.S_IFREG | 0444
		stat.st_nlink = 1
		stat.st_size = C.off_t(len(helloContent))
		entry.attr = stat

		C.fuse_reply_entry(req, &entry)
	} else {
		C.fuse_reply_err(req, C.ENOENT)
	}
}

//export go_ll_readdir
func go_ll_readdir(req C.fuse_req_t, ino C.fuse_ino_t, size C.size_t, off C.off_t, fi *C.struct_fuse_file_info) {
	if ino != rootInode {
		C.fuse_reply_err(req, C.ENOENT)
		return
	}

	buf := make([]byte, int(size))
	var written C.size_t

	if off == 0 {
		dot := C.CString(".")
		defer C.free(unsafe.Pointer(dot))
		var stat C.struct_stat
		stat.st_ino = rootInode
		written += C.fuse_add_direntry(req, (*C.char)(unsafe.Pointer(&buf[written])), size-written, dot, &stat, 1)
	}
	if off <= 1 {
		dotdot := C.CString("..")
		defer C.free(unsafe.Pointer(dotdot))
		var stat C.struct_stat
		stat.st_ino = rootInode
		written += C.fuse_add_direntry(req, (*C.char)(unsafe.Pointer(&buf[written])), size-written, dotdot, &stat, 2)
	}
	if off <= 2 {
		hello := C.CString(helloName)
		defer C.free(unsafe.Pointer(hello))
		var stat C.struct_stat
		stat.st_ino = helloInode
		written += C.fuse_add_direntry(req, (*C.char)(unsafe.Pointer(&buf[written])), size-written, hello, &stat, 3)
	}

	C.fuse_reply_buf(req, (*C.char)(unsafe.Pointer(&buf[0])), written)
}

//export go_ll_open
func go_ll_open(req C.fuse_req_t, ino C.fuse_ino_t, fi *C.struct_fuse_file_info) {
	if ino != helloInode {
		C.fuse_reply_err(req, C.ENOENT)
		return
	}
	C.fuse_reply_open(req, fi)
}

//export go_ll_read
func go_ll_read(req C.fuse_req_t, ino C.fuse_ino_t, size C.size_t, off C.off_t, fi *C.struct_fuse_file_info) {
	if ino != helloInode {
		C.fuse_reply_err(req, C.ENOENT)
		return
	}
	if off >= C.off_t(len(helloContent)) {
		C.fuse_reply_buf(req, nil, 0)
		return
	}
	end := int(off) + int(size)
	if end > len(helloContent) {
		end = len(helloContent)
	}
	C.fuse_reply_buf(req, (*C.char)(unsafe.Pointer(&helloContent[off])), C.size_t(end-int(off)))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <mountpoint>\n", os.Args[0])
		os.Exit(1)
	}

	mountpointStr := os.Args[len(os.Args)-1]
	mountpoint := C.CString(mountpointStr)
	defer C.free(unsafe.Pointer(mountpoint))

	cargs := make([]*C.char, len(os.Args)-1)
	for i, s := range os.Args[:len(os.Args)-1] {
		cargs[i] = C.CString(s)
	}
	defer func() {
		for _, s := range cargs {
			C.free(unsafe.Pointer(s))
		}
	}()
	argv := &cargs[0]
	args := C.struct_fuse_args{
		argc:      C.int(len(os.Args) - 1),
		argv:      argv,
		allocated: 0,
	}

	C.init_ops()

	ch := C.fuse_mount(mountpoint, &args)
	if ch == nil {
		os.Exit(1)
	}

	se := C.fuse_session_new(&args, &C.ll_ops, C.size_t(unsafe.Sizeof(C.ll_ops)), nil)
	if se == nil {
		C.fuse_unmount(mountpoint, ch)
		os.Exit(1)
	}

	if C.fuse_set_signal_handlers(se) != 0 {
		C.fuse_session_destroy(se)
		C.fuse_unmount(mountpoint, ch)
		os.Exit(1)
	}

	C.fuse_session_add_chan(se, ch)

	ret := C.fuse_session_loop(se)

	C.fuse_session_remove_chan(ch)
	C.fuse_remove_signal_handlers(se)
	C.fuse_session_destroy(se)
	C.fuse_unmount(mountpoint, ch)
	C.fuse_opt_free_args(&args)

	if ret != 0 {
		os.Exit(1)
	}
}
