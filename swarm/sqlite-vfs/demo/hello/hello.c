#include <sqlite3.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <stdio.h>
#include "hello.h"

#ifndef TEST_HELLO
#include "_cgo_export.h"
#endif

sqlite3 *bzz_dbh;
FILE *fdebug;
FILE *fnull;

void lasterr(int l, char *s) {
	int c;
	const char *err;

	err = sqlite3_errmsg(bzz_dbh);
	if (strlen(err) < l) {
		c = strlen(err);
	} else {
		c = l; 
	}
	strncpy(s, err, c);
}

int bzzClose(sqlite3_file *file) {
	fprintf(fdebug, "close!\n");
	return SQLITE_OK;
}

int bzzRead(sqlite3_file *file, void *p, int iAmt, sqlite3_int64 iOfst) {
	fprintf(fdebug, "read %d offset %d!\n", iAmt, iOfst);
#ifndef TEST_HELLO
	long long int c;
	bzzFile *bf;
       
	bf = (bzzFile*)file;
	c = GoBzzRead(bf->fd, p, iAmt, iOfst);
	fprintf(fdebug, "read return %d\n", c);
	if (c < 0) {
		return SQLITE_IOERR;
	} else if (c < iAmt) {
		return SQLITE_IOERR_SHORT_READ;
	}
#endif
	return SQLITE_OK;
}

int bzzWrite(sqlite3_file *file, const void *p, int iAmt, sqlite3_int64 iOfst) {
	fprintf(fdebug, "write!\n");
	return SQLITE_IOERR_WRITE;
}

int bzzTruncate(sqlite3_file *file, sqlite3_int64 size) {
	fprintf(fdebug, "trunc!\n");
	return SQLITE_IOERR_TRUNCATE;
}

int bzzSync(sqlite3_file *file, int flags) {
	fprintf(fdebug, "sync!\n");
	return SQLITE_IOERR_FSYNC;
}

int bzzFileSize(sqlite3_file *file, sqlite3_int64 *o_size) {
	fprintf(fdebug, "fsize!\n");
#ifndef TEST_HELLO
	bzzFile *bf = (bzzFile*)file;
	long long int s = GoBzzFileSize(bf->fd);
	if (s < 0) {
		return SQLITE_NOTFOUND;
	}
	*o_size = (sqlite3_int64)s;
#endif
	return SQLITE_OK;
}

int bzzLock(sqlite3_file *file, int l) {
	fprintf(fdebug, "lock!\n");
	//return SQLITE_IOERR_LOCK;
	return SQLITE_OK;
}

int bzzUnlock(sqlite3_file *file, int l) {
	fprintf(fdebug, "unlock!\n");
	//return SQLITE_IOERR_LOCK;
	return SQLITE_OK;
}

int bzzCheckReservedLock(sqlite3_file *file, int *o_res) {
	fprintf(fdebug, "chklock!\n");
	return SQLITE_IOERR_LOCK;
}

int bzzFileControl(sqlite3_file *file, int op, void *o_arg) {
	sqlite3_int64 *sz;
	switch (op) {
		case 18:
		       	sz = (sqlite3_int64*)o_arg;
			fprintf(fdebug, "fctrl mmapsize (%d): %d!\n", op, *sz);
			break;
		default:
			fprintf(fdebug, "fctrl %d!\n", op);
	}
	return SQLITE_OK;
}

int bzzSectorSize(sqlite3_file *file) {
	fprintf(fdebug, "secsize!\n");
	return BZZVFS_SECSIZE;
}

int bzzDeviceCharacteristics(sqlite3_file *file) {
	fprintf(fdebug, "devcharacteristics!\n");
	return SQLITE_IOCAP_ATOMIC4K | SQLITE_IOCAP_SEQUENTIAL;
}

int bzzShmMap(sqlite3_file *file, int iPg, int pgsz, int n, void volatile **v) {
	fprintf(fdebug, "shmmap!\n");
	return SQLITE_OK;
}

int bzzShmLock(sqlite3_file *file, int offset, int n, int flags) {
	fprintf(fdebug, "shmlock!\n");
	return SQLITE_OK;
}

void bzzShmBarrier(sqlite3_file *file) {
	fprintf(fdebug, "shmbarrier!\n");
	return;
}

int bzzShmUnmap(sqlite3_file *file, int deleteFlag) {
	fprintf(fdebug, "shmunmap!\n");
	return SQLITE_OK;
}

int bzzFetch(sqlite3_file *file, sqlite3_int64 iOfst, int iAmt, void **pp) {
	fprintf(fdebug, "fetch!\n");
	return SQLITE_OK;
}

int bzzUnfetch(sqlite3_file *file, sqlite3_int64 iOfst, void *p) {
	fprintf(fdebug, "unfetch!\n");
	return SQLITE_OK;
}

int bzzOpen(sqlite3_vfs *vfs, const char *zName, sqlite3_file *file, int flags, int *outflags) {
	fprintf(fdebug, "bzzopen %s\n", zName);
	static sqlite3_io_methods bzzIO = {
		3,
		bzzClose,
		bzzRead,
		bzzWrite,
		bzzTruncate,
		bzzSync,
		bzzFileSize,
		bzzLock,
		bzzUnlock,
		bzzCheckReservedLock,
		bzzFileControl,
		bzzSectorSize,
		bzzDeviceCharacteristics,
		bzzShmMap,
		bzzShmLock,
		bzzShmBarrier,
		bzzShmUnmap,
		bzzFetch,
		bzzUnfetch
	};
	bzzFile *bf = (bzzFile*)file;
	memset(bf, 0, sizeof(bzzFile));
#ifndef TEST_HELLO
	char name[strlen(zName)+1];
	strcpy(name, zName);
	if (GoBzzOpen(name, &bf->fd) > 0) {
		return SQLITE_NOTFOUND;	
	}
#endif
	bf->base.pMethods = &bzzIO;
	return SQLITE_OK;
}

int bzzDelete(sqlite3_vfs *vfs, const char *zName, int syncDir) {
	fprintf(fdebug, "del!\n");
	return SQLITE_OK;
}

int bzzAccess(sqlite3_vfs *vfs, const char *zName, int flags, int *pResOut) {
	fprintf(fdebug, "axx %d!\n", flags);
	*pResOut = SQLITE_OK;
	return SQLITE_OK;
}

int bzzFullPathname(sqlite3_vfs *vfs, const char *zName, int nOut, char *zOut) {
	fprintf(fdebug, "fullpathname!\n");
	size_t l = strlen(zName);
	if (nOut < l) {
		l = nOut;
	}
	strncpy(zOut, zName, l);
	return SQLITE_OK;
}

void* bzzDlOpen(sqlite3_vfs *vfs, const char *zName) {
	fprintf(fdebug, "dlopen!\n");
	return 0;
}

void bzzDlError(sqlite3_vfs *vfs, int nByte, char *zErrMsg) {
	fprintf(fdebug, "dlerr:");
	sqlite3_snprintf(nByte, zErrMsg, "not supported\0");
	return;
}

void (*bzzDlSym(sqlite3_vfs *vfs, void *foo, const char *zSymbol))(void) {
	fprintf(fdebug, "dlsym!\n");
	return 0;
}

void bzzDlClose(sqlite3_vfs *vfs, void *foo) {
	fprintf(fdebug, "dlclose!\n");
	return;
}

int bzzRandomness(sqlite3_vfs *vfs, int nByte, char *zByte) {
	fprintf(fdebug, "random!\n");
	return SQLITE_OK;
}

int bzzSleep(sqlite3_vfs *vfs, int us) {
	fprintf(fdebug, "sleep %d!\n", us);
	return us;
}

int bzzCurrentTime(sqlite3_vfs *vfs, double *o_t) {
	fprintf(fdebug, "curtime!\n");
	time_t t = time(0);
	o_t = (double*)&t;
	return SQLITE_OK;
}

int bzzGetLastError(sqlite3_vfs *vfs, int n, char *s) {
	fprintf(fdebug, "lasterr!\n");
	return SQLITE_OK;
}

int bzzCurrentTimeInt64(sqlite3_vfs *vfs, sqlite3_int64 *o_t) {
	fprintf(fdebug, "curtime64!\n");
	return SQLITE_OK;
}

int bzzSetSystemCall(sqlite3_vfs *vfs, const char *zName, sqlite3_syscall_ptr pSys) {
	fprintf(fdebug, "setsyscall!\n");
	return SQLITE_OK;
}

sqlite3_syscall_ptr bzzGetSystemCall(sqlite3_vfs *vfs, const char *zName) {
	fprintf(fdebug, "getsyscall!\n");
	return 0;
}

const char* bzzNextSystemCall(sqlite3_vfs *vfs, const char *zName) {
	fprintf(fdebug, "nextsyscall!\n");
	return "\0";
}

int bzzvfs_register() {
	int r;

	r = sqlite3_initialize();
	if (r != SQLITE_OK) {
		return r;
	}

	static sqlite3_vfs bzzvfs = {
		3,
		sizeof(bzzFile),
		BZZVFS_MAXPATHNAME,
		0,
		BZZVFS_ID,
		0,
		bzzOpen,
		bzzDelete,
		bzzAccess,
		bzzFullPathname,
		bzzDlOpen,
		bzzDlError,
		bzzDlSym,
		bzzDlClose,
		bzzRandomness,
		bzzSleep,
		bzzCurrentTime,
		bzzGetLastError,
		bzzCurrentTimeInt64,
		bzzSetSystemCall,
		bzzGetSystemCall,
		bzzNextSystemCall
	};
	fnull = fopen("/dev/null", "w");
	fdebug = fnull;
	return sqlite3_vfs_register(&bzzvfs, 0);
}

int bzzvfs_open(char *name) {
	return sqlite3_open_v2(name, &bzz_dbh, SQLITE_OPEN_READONLY, BZZVFS_ID);
}

void bzzvfs_close() {
	fclose(fnull);
	sqlite3_close_v2(bzz_dbh);
	sqlite3_shutdown();
}

int bzzvfs_exec(int sqlLen, const char *sql, int resLen, char *res) {
	int r;
	int i;
	char *clip;

	sqlite3_stmt *sth;
	r = sqlite3_prepare(bzz_dbh, sql, sqlLen, &sth, NULL);
	if (r != SQLITE_OK) {
		lasterr(resLen, res);
		return r;
	}

	clip = malloc(sizeof(char) * 9);
	i = 0;	
	while ((r = sqlite3_step(sth)) != SQLITE_DONE) {

		int id;
		const void *val;

		if (r != SQLITE_ROW) {
			lasterr(resLen, res);
			return r;
		}

		i++;
		id = sqlite3_column_int(sth, 0);
		val = sqlite3_column_blob(sth, 1);
		sprintf(clip, "%08x", *(unsigned int*)val);
		fprintf(stdout, ">>>> row %d: %d -> %s\n", i, id, clip);
	}
	free(clip);
	return 0;
}

void bzzvfs_debug(int b) {
	if (b) {
		fdebug = stderr;
	} else {
		fdebug = fnull;
	}
}
