// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"db";
	"fmt";
	"http";
	"os";
	"strconv";
)

// These constants can be or'd together and passed as the
// "flags" option to Open(). Some of them only apply if
// the "vfs" option is also passed. See SQLite documentation
// for details. Note that we always force OpenFullMutex,
// so passing OpenNoMutex has no effect. See also FlagURL().
const (
	OpenReadOnly		= 0x00000001;
	OpenReadWrite		= 0x00000002;
	OpenCreate		= 0x00000004;
	OpenDeleteOnClose	= 0x00000008;	// VFS only
	OpenExclusive		= 0x00000010;	// VFS only
	OpenMainDb		= 0x00000100;	// VFS only
	OpenTempDb		= 0x00000200;	// VFS only
	OpenTransientDb		= 0x00000400;	// VFS only
	OpenMainJournal		= 0x00000800;	// VFS only
	OpenTempJournal		= 0x00001000;	// VFS only
	OpenSubJournal		= 0x00002000;	// VFS only
	OpenMasterJournal	= 0x00004000;	// VFS only
	OpenNoMutex		= 0x00008000;
	OpenFullMutex		= 0x00010000;
	OpenSharedCache		= 0x00020000;
	OpenPrivateCache	= 0x00040000;
)

// Constants for sqlite3_config() used only internally.
// In fact only *one* is used. See SQLite documentation
// for details.
const (
	_	= iota;
	configSingleThread;
	configMultiThread;
	configSerialized;
	configMalloc;
	configGetMalloc;
	configScratch;
	configPageCache;
	configHeap;
	configMemStatus;
	configMutex;
	configGetMutex;
	_;
	configLookAside;
	configPCache;
	configGetPCache;
)

// after we run into a locked database/table,
// we'll retry for this long
const defaultTimeoutMilliseconds = 16 * 1000

// SQLite version information
var Version db.VersionSignature
// SQLite connection factory
var Open db.OpenSignature

func init() {
	// Idiom to ensure that we actually conform
	// to the database API for functions as well.
	Version = version;
	Open = open;

	// Supposedly serialized mode is the default,
	// but let's make sure...
	rc := sqlConfig(configSerialized);
	if rc != StatusOk {
		panic("sqlite3 fatal error: can't switch to serialized mode")
	}
}

// The SQLite database interface returns keys "version",
// "sqlite3.sourceid", and "sqlite3.versionnumber"; the
// latter are specific to SQLite.
func version() (data map[string]string, error os.Error) {
	// TODO: fake client and server keys?
	data = make(map[string]string);
	data["version"] = sqlVersion();
	i := sqlVersionNumber();
	data["sqlite3.versionnumber"] = strconv.Itob(i, 10);
	data["sqlite3.sourceid"] = sqlSourceId();
	return;
}

func parseConnInfo(str string) (name string, flags int, vfs string, error os.Error) {
	var url *http.URL;

	url, error = http.ParseURL(str);
	if error != nil {
		return	// XXX really return error from ParseURL?
	}

	if len(url.Scheme) > 0 {
		if url.Scheme != "sqlite3" {
			error = &DriverError{fmt.Sprintf("Open: unknown scheme %s expected sqlite3", url.Scheme)};
			return;
		}
	}

	if len(url.Path) == 0 {
		error = &DriverError{"Open: no path or database name"};
		return;
	} else {
		name = url.Path
	}

	if len(url.RawQuery) > 0 {
		options, e := db.ParseQueryURL(url.RawQuery);
		if e != nil {
			error = e;
			return	// XXX really return error from ParseQueryURL?
		}
		rflags, ok := options["flags"];
		if ok {
			flags, error = strconv.Atoi(rflags);
			if error != nil {
				return	// XXX really return error from Atoi?
			}
		}
		vfs, ok = options["vfs"];
	}

	return;
}

func open(url string) (connection db.Connection, error os.Error) {
	var name string;
	var flags int;
	var vfs string;

	name, flags, vfs, error = parseConnInfo(url);
	if error != nil {
		return
	}

	// We want all connections to be in serialized threading
	// mode, so we fiddle with the flags to make sure.
	flags &^= OpenNoMutex;
	flags |= OpenFullMutex;

	conn := new(Connection);
	var rc int;
	conn.handle, rc = sqlOpen(name, flags, vfs);

	if rc != StatusOk {
		error = conn.error();
		// did we get a handle anyway? if so we need to
		// close it, but that could trigger another,
		// secondary error; for now we ignore that one
		if conn.handle != nil {
			_ = conn.Close();
		}
		return;
	}

	rc = conn.handle.sqlBusyTimeout(defaultTimeoutMilliseconds);
	if rc != StatusOk {
		error = conn.error();
		// ignore potential secondary error
		_ = conn.Close();
		return;
	}

	rc = conn.handle.sqlExtendedResultCodes(true);
	if rc != StatusOk {
		error = conn.error();
		// ignore potential secondary error
		_ = conn.Close();
		return;
	}

	connection = conn;
	return;
}

/*
func (self *Cursor) FetchRow() (data map[string]interface{}, error os.Error) {
	if !self.result {
		error = &DriverError{"FetchRow: No results to fetch!"};
		return;
	}

	nColumns := int(C.sqlite3_column_count(self.statement.handle));
	if nColumns <= 0 {
		error = &DriverError{"FetchRow: No columns in result!"};
		return;
	}

	data = make(map[string]interface{}, nColumns);
	for i := 0; i < nColumns; i++ {
		text := C.wsq_column_text(self.statement.handle, C.int(i));
		name := C.wsq_column_name(self.statement.handle, C.int(i));
		data[C.GoString(name)] = C.GoString(text);
	}

	// try to get another row
	rc := C.sqlite3_step(self.statement.handle);

	if rc != StatusDone && rc != StatusRow {
		// presumably any other outcome is an error
		error = self.connection.error()
	}

	if rc == StatusDone {
		self.result = false;
		// clean up when done
		self.statement.clear();
	}

	return;
}
*/
