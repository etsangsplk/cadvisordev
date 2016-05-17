
echo copying wavefront storage driver from cadvisor codebase to wfsrc
cp share/src/github.com/google/cadvisor/storage/wavefront/wavefront.go share/wfsrc/storage/wavefront/
cp share/src/github.com/google/cadvisor/storagedriver.go share/wfsrc/

echo working files copied to wfsrc, check into git when ready.
