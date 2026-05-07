package lb

// TODO: impl least connection attempts:
// approach 1: wrap AfterConnect: incr weights for ok attempts (conn.Conn().RemoteAddr().String())
// approach 2: (wrap LookupFunc: store host->ips mapping)  + (wrap DialFunc: ip->incr weight for ok attempts)
