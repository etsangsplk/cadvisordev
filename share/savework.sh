
echo copying wavefront storage driver from cadvisor codebase to wfsrc
cp src/github.com/google/cadvisor/storage/wavefront/wavefront.go wfsrc/storage/wavefront/
cp src/github.com/google/cadvisor/storagedriver.go wfsrc/

echo working files copied to wfsrc, check into git when ready.
