go test -coverpkg=./... -coverprofile report.out -covermode=atomic ./... &&\
grep -v -E "mock|gen|test|cmd|pb|interceptors" report.out > buf && mv buf report.out &&\
go tool cover -func=report.out &&\
go tool cover -html=report.out
