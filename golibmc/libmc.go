package libmc

/*
#cgo  CFLAGS: -I ${SRCDIR}/../include
#cgo LDFLAGS: -L ${SRCDIR}/../build -l mc -l stdc++
#include "c_client.h"

*/
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

const (
	MC_HASH_MD5      = 0
	MC_HASH_FNV1_32  = 1
	MC_HASH_FNV1A_32 = 2
	MC_HASH_CRC_32   = 3
)
const MC_DEFAULT_PORT = 11211

type Client struct {
	_imp    unsafe.Pointer
	servers []string
	prefix  string
	noreply bool
}

func (self *Client) Init(servers []string, noreply bool, prefix string,
	hash_fn string, failover bool) {
	// TODO handle hash_fn
	self._imp = C.client_create()

	n := len(servers)
	c_hosts := make([]*C.char, n)
	c_ports := make([]C.uint32_t, n)
	c_aliases := make([]*C.char, n)

	for i, srv := range servers {
		addr_alias := strings.Split(srv, ":")

		addr := addr_alias[0]
		if len(addr_alias) == 2 {
			c_alias := C.CString(addr_alias[1])
			defer C.free(unsafe.Pointer(c_alias))
			c_aliases[i] = c_alias
		}

		host_port := strings.Split(addr, ":")
		host := host_port[0]
		var c_host *C.char = C.CString(host)
		defer C.free(unsafe.Pointer(c_host))
		c_hosts[i] = c_host

		if len(host_port) == 2 {
			port, err := strconv.Atoi(host_port[1])
			if err != nil {
				fmt.Println(err) // TODO handle error
			}
			c_ports[i] = C.uint32_t(port)
		} else {
			c_ports[i] = C.uint32_t(MC_DEFAULT_PORT)
		}
	}

	failoverInt := 0
	if failover {
		failoverInt = 1
	}

	C.client_init(
		self._imp,
		(**C.char)(unsafe.Pointer(&c_hosts[0])),
		(*C.uint32_t)(unsafe.Pointer(&c_ports[0])),
		C.size_t(n),
		(**C.char)(unsafe.Pointer(&c_aliases[0])),
		C.int(failoverInt),
	)

	self.prefix = prefix
	self.noreply = noreply
}

func (self *Client) Destroy() {
	C.client_destroy(self._imp)
}

func (self *Client) Version() (map[string]string, error) {
	var rst *C.broadcast_result_t
	var n C.size_t
	rv := make(map[string]string)

	err_code := C.client_version(self._imp, &rst, &n)
	defer C.client_destroy_broadcast_result(self._imp)
	sr := unsafe.Sizeof(*rst)

	for i := 0; i < int(n); i += 1 {
		if rst.lines == nil || rst.line_lens == nil {
			continue
		}

		host := C.GoString(rst.host)
		version := C.GoStringN(*rst.lines, C.int(*rst.line_lens))
		rv[host] = version
		rst = (*C.broadcast_result_t)(unsafe.Pointer(uintptr(unsafe.Pointer(rst)) + sr))
	}

	if err_code != 0 {
		return rv, errors.New(strconv.Itoa(int(err_code)))
	}

	return rv, nil
}
