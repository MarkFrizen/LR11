use nix::fcntl::{flock, FlockArg};
use std::fs::OpenOptions;
use std::io::{self, Read, Write};
use std::os::unix::io::AsRawFd;
use std::path::PathBuf;

use crate::storage::LogStorage;

/// Файловая реализация LogStorage (SRP — только файловые операции)
pub struct FileLogStorage {
    filepath: PathBuf,
}

impl FileLogStorage {
    pub fn new(filepath: &str) -> Self {
        Self {
            filepath: PathBuf::from(filepath),
        }
    }
}

impl LogStorage for FileLogStorage {
    fn write_line(&self, line: &str) -> Result<(), io::Error> {
        let mut f = OpenOptions::new()
            .create(true)
            .append(true)
            .open(&self.filepath)?;

        let fd = f.as_raw_fd();
        flock(fd, FlockArg::LockExclusive).map_err(|e| io::Error::new(io::ErrorKind::Other, e.to_string()))?;

        f.write_all(line.as_bytes())?;

        let _ = flock(fd, FlockArg::Unlock);
        Ok(())
    }

    fn read_all(&self) -> Result<String, io::Error> {
        let mut file = match OpenOptions::new().read(true).open(&self.filepath) {
            Ok(f) => f,
            Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(String::new()),
            Err(e) => return Err(e),
        };

        let fd = file.as_raw_fd();
        flock(fd, FlockArg::LockShared).map_err(|e| io::Error::new(io::ErrorKind::Other, e.to_string()))?;

        let mut content = String::new();
        file.read_to_string(&mut content)?;

        let _ = flock(fd, FlockArg::Unlock);
        Ok(content)
    }
}
