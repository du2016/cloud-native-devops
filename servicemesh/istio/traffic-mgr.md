A-B-bar.foo.com

SERVICE A
-----------
SERVICE B(VS)
-----------
ROUTE(timeout,retry,fault)
-----------
Satic sourcelabel header weight uri
---------------
DS
--------
subsets (trafficpolicy<maxconn|>)
--------
Service entry
---------
VS()
