"""Export versionable files for a local scan without traversing ignored secrets/caches."""
from pathlib import Path
import shutil
import subprocess
import sys
import re

root = Path(__file__).resolve().parents[2]
name = sys.argv[1] if len(sys.argv) == 2 else 'security-current-source'
if not re.fullmatch(r'[a-z0-9-]+', name):
    raise SystemExit('Choose a plain private export directory name.')
destination = (root / '.local' / name).resolve()
result = subprocess.run(['git', 'ls-files', '--cached', '--others', '--exclude-standard', '-z'],
                        cwd=root, check=True, capture_output=True)
if destination.exists():
    raise SystemExit('Scan export already exists; choose a new isolated export before repeating.')
count = 0
for value in result.stdout.decode('utf-8').split('\0'):
    if not value:
        continue
    source = root / value
    if not source.is_file() or source.is_symlink():
        continue
    # Git output is untrusted input: never copy outside this workspace/export.
    if not source.resolve().is_relative_to(root) or not (destination / value).resolve().is_relative_to(destination):
        raise SystemExit('Unsafe source path in Git inventory')
    target = destination / value
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, target)
    count += 1
print(f'Exported {count} versionable files for a redacted local secret scan.')
