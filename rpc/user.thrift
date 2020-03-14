/**
 *  bool        Boolean, one byte
 *  i8 (byte)   Signed 8-bit integer
 *  i16         Signed 16-bit integer
 *  i32         Signed 32-bit integer
 *  i64         Signed 64-bit integer
 *  double      64-bit floating point value
 *  string      String
 *  binary      Blob (byte array)
 *  map<t1,t2>  Map from one type to another
 *  list<t1>    Ordered list of one type
 *  set<t1>     Set of unique elements of one type
 */

namespace java user
namespace php user


struct UserStruct {
  1: i64 id = 0,
  2: string username,
  3: string nickname,
  4: i32 state,
  5: i64 create_at
  6: i64 update_at
}

service UserService {
   UserStruct reg(1:string username, 2:string nickname, 3:string password),
   UserStruct login(1:string username, 2:string password),
}

