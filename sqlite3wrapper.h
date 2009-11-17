#ifndef _WRAP_SQLITE3_H_
#define _WRAP_SQLITE3_H_

/*
	Apparently cgo has problems with some #include files,
	including (hah!) sqlite3.h and friends. So we need a
	simple wrapper that hides the real sqlite3.h for now.

	Note that this wrapper is incomplete for now. I just
	needed something that works at all.
*/

/* really sqlite3 */
typedef void *wsq_db;
/* really sqlite3_stmt */
typedef void *wsq_st;

/* really sqlite3_open_v2 */
int wsq_open(const char *name, wsq_db *database, int flags, const char *vfs);

/* really sqlite3_prepare_v2 */
int wsq_prepare(wsq_db database, const char *sql, int length, wsq_st *statement, const char **tail);

/* really sqlite3_step */
int wsq_step(wsq_st statement);

/* really sqlite3_column_count */
int wsq_column_count(wsq_st statement);

/* really sqlite3_column_type */
int wsq_column_type(wsq_st statement, int column);

/* really sqlite3_column_name */
const char *wsq_column_name(wsq_st statement, int column);

/* really sqlite3_column_text */
const unsigned char *wsq_column_text(wsq_st statement, int column);

/* really sqlite3_finalize */
int wsq_finalize(wsq_st statement);

/* really sqlite3_close */
int wsq_close(wsq_db database);

#endif /* !_WRAP_SQLITE3_H_ */
