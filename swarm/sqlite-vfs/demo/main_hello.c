#include <sqlite3.h>
#include "hello.h"

int main(int argc, char **argv) {
	if (bzzvfs_register() != SQLITE_OK) {
		return 1;
	}
	bzzvfs_debug(1);
	sqlite3_vfs *thevfs = sqlite3_vfs_find(BZZVFS_ID);
	if (thevfs == 0) {
		return 1;
	}
	return bzzvfs_open("abcdef0123456789abcdef1234567890abcdef1234567890abcdef1234568790");
}
