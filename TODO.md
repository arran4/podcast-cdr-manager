# TODOs

- In `disk.go`: `createDiskFilename` and `createDiskIsoName` has negative index panics and panics with large inputs causing overflows, fix by using type casting properly.
- In `cmd/podcastcdrmanager/diskNext.go`: make it get the size some-other way, until then hard fail.
