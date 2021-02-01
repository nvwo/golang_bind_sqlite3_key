package main

/*
#include <stdio.h>
#include "sqlite3.h"

#define ABORT 0
#define CONTINUE 1

void* sqlite3_sleep_go;
typedef int (*sleep_func)(int);

int busy_handler(void *data, int attempt) {
    printf("attempt: %d\n", attempt);
    if (attempt < 10) {
        ((sleep_func)sqlite3_sleep_go)(1000);
        return CONTINUE;
    }
    return ABORT;
}
*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func basePath() (string) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}

	return dir + string(os.PathSeparator)
}

const (
	SQLITE_OK = 0
	SQLITE_BUSY = 5
	SQLITE_MISUSE  =    21
	SQLITE_ROW = 100

	SQLITE_OPEN_READONLY   =      0x00000001
)

var (
	modSQLite3                              *syscall.LazyDLL
	dll_sqlite3_open                     *syscall.LazyProc
	dll_sqlite3_open_v2             *syscall.LazyProc
	dll_sqlite3_sleep               *syscall.LazyProc
	dll_sqlite3_busy_handler      *syscall.LazyProc
	dll_sqlite3_key               *syscall.LazyProc
	dll_sqlite3_exec               *syscall.LazyProc
	dll_sqlite3_close                   *syscall.LazyProc
	dll_sqlite3_prepare                   *syscall.LazyProc
	dll_sqlite3_errmsg                   *syscall.LazyProc
	dll_sqlite3_step                   *syscall.LazyProc
	dll_sqlite3_finalize                *syscall.LazyProc
	dll_sqlite3_column_int                   *syscall.LazyProc
	dll_sqlite3_column_text                    *syscall.LazyProc
)

func Uint(i int)uint{
	return *(*uint)(unsafe.Pointer(&i))
}

func main(){
	dllName := "sqlite3.dll"
	filePath := basePath() + dllName
	if exist, _ := exists(filePath); exist {
		modSQLite3 = syscall.NewLazyDLL(filePath)
	}else{
		return
	}
	dll_sqlite3_open  = modSQLite3.NewProc("sqlite3_open")
	dll_sqlite3_open_v2   = modSQLite3.NewProc("sqlite3_open_v2")
	dll_sqlite3_sleep  = modSQLite3.NewProc("sqlite3_sleep")
	dll_sqlite3_busy_handler = modSQLite3.NewProc("sqlite3_busy_handler")
	dll_sqlite3_key  = modSQLite3.NewProc("sqlite3_key")
	dll_sqlite3_exec  = modSQLite3.NewProc("sqlite3_exec")
	dll_sqlite3_close = modSQLite3.NewProc("sqlite3_close")
	dll_sqlite3_prepare   = modSQLite3.NewProc("sqlite3_prepare")
	dll_sqlite3_errmsg   = modSQLite3.NewProc("sqlite3_errmsg")
	dll_sqlite3_step   = modSQLite3.NewProc("sqlite3_step")
	dll_sqlite3_finalize  = modSQLite3.NewProc("sqlite3_finalize")
	dll_sqlite3_column_int  = modSQLite3.NewProc("sqlite3_column_int")
	dll_sqlite3_column_text  = modSQLite3.NewProc("sqlite3_column_text")

	var db uintptr
	s := "./test.db"
	b := append([]byte(s), 0)
	//rc,_,_ := dll_sqlite3_open.Call(uintptr(unsafe.Pointer(&b[0])),uintptr(unsafe.Pointer(&db)))
	rc,_,_ := dll_sqlite3_open.Call(uintptr(unsafe.Pointer(&b[0])),uintptr(unsafe.Pointer(&db)),SQLITE_OPEN_READONLY,0)
	if int(rc)!=SQLITE_OK {
		log.Fatalf("dll_sqlite3_open failed\n")
		return
	}
	C.sqlite3_sleep_go = unsafe.Pointer(dll_sqlite3_sleep)
	dll_sqlite3_busy_handler.Call(uintptr(unsafe.Pointer(db)), uintptr(unsafe.Pointer(C.busy_handler)), 0);

	defer func() {
		rc,_,_ = dll_sqlite3_close.Call(uintptr(unsafe.Pointer(db)))
		if int(rc)!=SQLITE_OK {
			rc,_,_ := dll_sqlite3_errmsg.Call(uintptr(unsafe.Pointer(db)))
			fmt.Printf("Error: %s.\n", string(rc));
			log.Fatalf("dll_sqlite3_close failed\n")
			return
		}
	}()

	key := "2DD29CA851E7B56E4697B0E1F08507293D761A05CE4D1B628663F411A8086D99"
	b = append([]byte(key), 0)
	rc,_,_ = dll_sqlite3_key.Call(uintptr(unsafe.Pointer(db)), uintptr(unsafe.Pointer(&b[0])), uintptr(len(key)))
	if int(rc)!=SQLITE_OK {
		log.Fatalf("dll_sqlite3_key failed\n")
		return
	}
	var result uintptr;

	// Consulta a realizar sobre la tabla.
	// En este caso quiero los campos idEmpresa y Nombre de la tabla Empresa
	key = "SELECT * FROM tb_record"
	b = append([]byte(key), 0)
	rc,_,_ = dll_sqlite3_prepare.Call(uintptr(unsafe.Pointer(db)), uintptr(unsafe.Pointer(&b[0])), uintptr(Uint(-1)), uintptr(unsafe.Pointer(&result)), uintptr(0));
	// Compruebo que no hay error
	if (rc != SQLITE_OK) {
		rc,_,_ := dll_sqlite3_errmsg.Call(uintptr(unsafe.Pointer(db)))
		fmt.Printf("Error: %s.\n", string(rc));
		return;
	}

	// Bucle de presentación en pantalla del resultado de la consulta

	for ;; {
		 rc,_,_ = dll_sqlite3_step.Call(uintptr(unsafe.Pointer(result)))
		 if int(rc) != SQLITE_ROW {
		 	break;
		 }
		l0,_,_ := dll_sqlite3_column_int.Call(uintptr(unsafe.Pointer(result)), uintptr(0))
		l1,_,_ := dll_sqlite3_column_text.Call(uintptr(unsafe.Pointer(result)), uintptr(1))
		fmt.Printf("El Id y nombre de la empresa son:  %d - %s.\n", int(l0), C.GoString((*C.char)(unsafe.Pointer(l1))));
	}

	dll_sqlite3_finalize.Call(uintptr(unsafe.Pointer(result)));



	//query := "SELECT * FROM tb_record"
	//b = append([]byte(query), 0)
	//rc,_,_ = dll_sqlite3_exec.Call(uintptr(unsafe.Pointer(db)), uintptr(unsafe.Pointer(&b[0])), uintptr(len(key)))
	//if int(rc)!=SQLITE_OK {
	//	log.Fatalf("dll_sqlite3_exec failed\n")
	//	return
	//}



}
