@echo off
setlocal EnableExtensions EnableDelayedExpansion
cd /d "%~dp0"

set "BASE_URL=http://localhost:8080"
set "OWNER_LOGIN=Owner"
set "OWNER_PASSWORD=123456"
set "DB_PATH=%CD%\data\storage.db"
set "PARALLEL_NAME=ZTEST_1_4"

if not exist "%DB_PATH%" (
  echo [ERR] DB file not found: %DB_PATH%
  exit /b 1
)

set "PY_CMD="
where python >nul 2>nul && set "PY_CMD=python"
if not defined PY_CMD (
  where py >nul 2>nul && set "PY_CMD=py -3"
)

if not defined PY_CMD (
  echo [ERR] Python not found. Install Python 3 or add it to PATH.
  exit /b 1
)

set "TMP_PY=%TEMP%\cspirt_parallel_test.py"

> "%TMP_PY%" echo import json, os, sqlite3, sys, urllib.request, urllib.error
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo BASE_URL = os.environ.get("BASE_URL", "http://localhost:8080")
>>"%TMP_PY%" echo DB_PATH = os.environ.get("DB_PATH", "data/storage.db")
>>"%TMP_PY%" echo OWNER_LOGIN = os.environ.get("OWNER_LOGIN", "Owner")
>>"%TMP_PY%" echo OWNER_PASSWORD = os.environ.get("OWNER_PASSWORD", "123456")
>>"%TMP_PY%" echo PARALLEL_NAME = os.environ.get("PARALLEL_NAME", "ZTEST_1_4")
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo TEST_CLASSES = ["ZP1A", "ZP1B", "ZP2A", "ZP2B"]
>>"%TMP_PY%" echo HELPERS = [
>>"%TMP_PY%" echo     ("zp_teacher1", "Teacher1", "One", "ZP1A"),
>>"%TMP_PY%" echo     ("zp_teacher2", "Teacher2", "Two", "ZP1B"),
>>"%TMP_PY%" echo     ("zp_teacher3", "Teacher3", "Three", "ZP2A"),
>>"%TMP_PY%" echo     ("zp_teacher4", "Teacher4", "Four", "ZP2B"),
>>"%TMP_PY%" echo ]
>>"%TMP_PY%" echo STUDENTS = [
>>"%TMP_PY%" echo     ("zp_student1", "Student1", "One", "ZP1A"),
>>"%TMP_PY%" echo     ("zp_student2", "Student2", "Two", "ZP1B"),
>>"%TMP_PY%" echo     ("zp_student3", "Student3", "Three", "ZP2A"),
>>"%TMP_PY%" echo     ("zp_student4", "Student4", "Four", "ZP2B"),
>>"%TMP_PY%" echo ]
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def api(method, path, payload=None, token=None):
>>"%TMP_PY%" echo     data = None if payload is None else json.dumps(payload).encode("utf-8")
>>"%TMP_PY%" echo     req = urllib.request.Request(BASE_URL + path, data=data, method=method)
>>"%TMP_PY%" echo     req.add_header("Content-Type", "application/json")
>>"%TMP_PY%" echo     if token:
>>"%TMP_PY%" echo         req.add_header("Authorization", "Bearer " + token)
>>"%TMP_PY%" echo     try:
>>"%TMP_PY%" echo         with urllib.request.urlopen(req, timeout=30) as resp:
>>"%TMP_PY%" echo             return resp.status, resp.read().decode("utf-8", "replace")
>>"%TMP_PY%" echo     except urllib.error.HTTPError as e:
>>"%TMP_PY%" echo         return e.code, e.read().decode("utf-8", "replace")
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def must_ok(method, path, payload=None, token=None):
>>"%TMP_PY%" echo     status, raw = api(method, path, payload, token)
>>"%TMP_PY%" echo     print(f"[{status}] {method} {path}")
>>"%TMP_PY%" echo     print(raw)
>>"%TMP_PY%" echo     print()
>>"%TMP_PY%" echo     if status >= 400:
>>"%TMP_PY%" echo         sys.exit(1)
>>"%TMP_PY%" echo     return raw
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def seed_db():
>>"%TMP_PY%" echo     conn = sqlite3.connect(DB_PATH)
>>"%TMP_PY%" echo     conn.execute("PRAGMA foreign_keys = ON")
>>"%TMP_PY%" echo     cur = conn.cursor()
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo     for login, _, _, _ in HELPERS:
>>"%TMP_PY%" echo         cur.execute("DELETE FROM users WHERE Login = ?", (login,))
>>"%TMP_PY%" echo     for login, _, _, _ in STUDENTS:
>>"%TMP_PY%" echo         cur.execute("DELETE FROM users WHERE Login = ?", (login,))
>>"%TMP_PY%" echo     for name in TEST_CLASSES:
>>"%TMP_PY%" echo         cur.execute("DELETE FROM classes WHERE Name = ?", (name,))
>>"%TMP_PY%" echo     cur.execute("DELETE FROM parallels WHERE Name = ?", (PARALLEL_NAME,))
>>"%TMP_PY%" echo     conn.commit()
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo     for name in TEST_CLASSES:
>>"%TMP_PY%" echo         cur.execute("""
>>"%TMP_PY%" echo             INSERT INTO classes
>>"%TMP_PY%" echo                 (Name, TeacherLogin, Members, Parallel,
>>"%TMP_PY%" echo                  FirstQuarterComplete, SecondQuarterComplete, ThirdQuarterComplete, QuarterComplete,
>>"%TMP_PY%" echo                  UserTotalRating, ClassTotalRating)
>>"%TMP_PY%" echo             VALUES (?, NULL, '[]', 0, 0, 0, 0, 0, 0, 0)
>>"%TMP_PY%" echo         """, (name,))
>>"%TMP_PY%" echo     conn.commit()
>>"%TMP_PY%" echo     return conn
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def login_owner():
>>"%TMP_PY%" echo     status, raw = api("POST", "/login", {"Login": OWNER_LOGIN, "Password": OWNER_PASSWORD})
>>"%TMP_PY%" echo     print(f"[{status}] POST /login")
>>"%TMP_PY%" echo     print(raw)
>>"%TMP_PY%" echo     print()
>>"%TMP_PY%" echo     if status != 200:
>>"%TMP_PY%" echo         sys.exit(1)
>>"%TMP_PY%" echo     return json.loads(raw)["accessToken"]
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def create_users(token):
>>"%TMP_PY%" echo     for login, name, last, cls in HELPERS:
>>"%TMP_PY%" echo         must_ok("PATCH", "/api/user/add", {
>>"%TMP_PY%" echo             "Name": name,
>>"%TMP_PY%" echo             "LastName": last,
>>"%TMP_PY%" echo             "FullName": [{"Name": name, "LastName": last}],
>>"%TMP_PY%" echo             "Login": login,
>>"%TMP_PY%" echo             "Password": "123456",
>>"%TMP_PY%" echo             "Role": "helper",
>>"%TMP_PY%" echo             "Class": cls
>>"%TMP_PY%" echo         }, token)
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo     for login, name, last, cls in STUDENTS:
>>"%TMP_PY%" echo         must_ok("PATCH", "/api/user/add", {
>>"%TMP_PY%" echo             "Name": name,
>>"%TMP_PY%" echo             "LastName": last,
>>"%TMP_PY%" echo             "FullName": [{"Name": name, "LastName": last}],
>>"%TMP_PY%" echo             "Login": login,
>>"%TMP_PY%" echo             "Password": "123456",
>>"%TMP_PY%" echo             "Role": "user",
>>"%TMP_PY%" echo             "Class": cls
>>"%TMP_PY%" echo         }, token)
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def get_class_ids(conn):
>>"%TMP_PY%" echo     cur = conn.cursor()
>>"%TMP_PY%" echo     placeholders = ",".join(["?"] * len(TEST_CLASSES))
>>"%TMP_PY%" echo     cur.execute(f"SELECT Id, Name FROM classes WHERE Name IN ({placeholders})", TEST_CLASSES)
>>"%TMP_PY%" echo     rows = cur.fetchall()
>>"%TMP_PY%" echo     id_by_name = {name: class_id for class_id, name in rows}
>>"%TMP_PY%" echo     missing = [name for name in TEST_CLASSES if name not in id_by_name]
>>"%TMP_PY%" echo     if missing:
>>"%TMP_PY%" echo         raise RuntimeError("Missing classes in DB: " + ", ".join(missing))
>>"%TMP_PY%" echo     return [id_by_name[name] for name in TEST_CLASSES]
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def create_parallel(token, class_ids):
>>"%TMP_PY%" echo     status, raw = api("PATCH", "/api/classes/parallel/add", {
>>"%TMP_PY%" echo         "Name": PARALLEL_NAME,
>>"%TMP_PY%" echo         "ClassesIds": class_ids
>>"%TMP_PY%" echo     }, token)
>>"%TMP_PY%" echo     print(f"[{status}] PATCH /api/classes/parallel/add")
>>"%TMP_PY%" echo     print(raw)
>>"%TMP_PY%" echo     print()
>>"%TMP_PY%" echo     if status >= 400:
>>"%TMP_PY%" echo         sys.exit(1)
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def get_parallel_id(conn):
>>"%TMP_PY%" echo     cur = conn.cursor()
>>"%TMP_PY%" echo     cur.execute("SELECT Id FROM parallels WHERE Name = ?", (PARALLEL_NAME,))
>>"%TMP_PY%" echo     row = cur.fetchone()
>>"%TMP_PY%" echo     if not row:
>>"%TMP_PY%" echo         raise RuntimeError("Parallel was not created in DB")
>>"%TMP_PY%" echo     return row[0]
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def smoke_tests(token, parallel_id):
>>"%TMP_PY%" echo     for method, path in [
>>"%TMP_PY%" echo         ("GET", "/api/classes"),
>>"%TMP_PY%" echo         ("GET", "/api/classes/parallel"),
>>"%TMP_PY%" echo         ("GET", f"/api/classes/parallel/{parallel_id}"),
>>"%TMP_PY%" echo         ("GET", f"/api/classes/parallel/{parallel_id}/users"),
>>"%TMP_PY%" echo         ("GET", f"/api/classes/parallel/{parallel_id}/best"),
>>"%TMP_PY%" echo         ("PATCH", f"/api/classes/quarter/complete?parallel_class_id={parallel_id}"),
>>"%TMP_PY%" echo     ]:
>>"%TMP_PY%" echo         status, raw = api(method, path, None, token)
>>"%TMP_PY%" echo         print(f"[{status}] {method} {path}")
>>"%TMP_PY%" echo         print(raw)
>>"%TMP_PY%" echo         print()
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo def main():
>>"%TMP_PY%" echo     print("Seed DB: " + DB_PATH)
>>"%TMP_PY%" echo     conn = seed_db()
>>"%TMP_PY%" echo     token = login_owner()
>>"%TMP_PY%" echo     create_users(token)
>>"%TMP_PY%" echo     class_ids = get_class_ids(conn)
>>"%TMP_PY%" echo     print("Class IDs:", class_ids)
>>"%TMP_PY%" echo     print()
>>"%TMP_PY%" echo     create_parallel(token, class_ids)
>>"%TMP_PY%" echo     parallel_id = get_parallel_id(conn)
>>"%TMP_PY%" echo     print("Parallel ID:", parallel_id)
>>"%TMP_PY%" echo     print()
>>"%TMP_PY%" echo     smoke_tests(token, parallel_id)
>>"%TMP_PY%" echo     conn.close()
>>"%TMP_PY%" echo.
>>"%TMP_PY%" echo if __name__ == "__main__":
>>"%TMP_PY%" echo     main()

%PY_CMD% "%TMP_PY%"
set "RC=%ERRORLEVEL%"

del "%TMP_PY%" >nul 2>nul
exit /b %RC%