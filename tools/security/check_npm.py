"""Fail on new high/critical root advisories, expired or shipped exceptions."""
from datetime import date
from pathlib import Path
import json
import os
import subprocess

ROOT = Path(__file__).resolve().parents[2]


def check(audit, lock, runtime_names, exceptions, today):
    if audit.get('auditReportVersion') != 2 or 'error' in audit or 'vulnerabilities' not in audit:
        raise ValueError('npm audit did not return a complete supported report')
    failures, accepted = [], []
    vulnerabilities = audit['vulnerabilities']
    for name, finding in vulnerabilities.items():
        for via in finding['via']:
            if isinstance(via, str):
                if via not in vulnerabilities:
                    failures.append(f'{name}: unresolved advisory chain')
                continue
            if via['severity'] not in {'high', 'critical'}:
                continue
            advisory = via['url'].rsplit('/', 1)[-1]
            versions = {lock['packages'][node]['version'] for node in finding['nodes']}
            match = next((entry for entry in exceptions if entry['module'] == name
                          and entry['advisory'] == advisory
                          and versions <= set(entry['versions'])), None)
            label = f'{name}: {advisory} ({", ".join(sorted(versions))})'
            if (not match or date.fromisoformat(match['expires']) < today
                    or not match.get('owner') or not match.get('reason') or not versions
                    or name in runtime_names):
                failures.append(label)
            else:
                accepted.append(label + ' exception expires ' + match['expires'])
    return failures, accepted


def main():
    web = ROOT / 'frontend'
    output = web / '.output/server'
    if not (output / 'package.json').is_file():
        raise SystemExit('Build the production frontend before reviewing npm exceptions.')
    runtime_names = set()
    manifests = [output / 'package.json', *output.glob('node_modules/**/package.json')]
    for path in manifests:
        manifest = json.loads(path.read_text(encoding='utf-8'))
        runtime_names.add(manifest.get('name', ''))
        runtime_names.update(manifest.get('dependencies', {}))
    command = 'npm.cmd' if os.name == 'nt' else 'npm'
    result = subprocess.run([command, 'audit', '--json'], cwd=web, capture_output=True, text=True, encoding='utf-8')
    if result.returncode not in (0, 1):
        raise SystemExit('npm audit could not complete; refusing a successful security result.')
    audit = json.loads(result.stdout)
    lock = json.loads((web / 'package-lock.json').read_text(encoding='utf-8'))
    exceptions = json.loads(Path(__file__).with_name('npm-exceptions.json').read_text(encoding='utf-8'))
    failures, accepted = check(audit, lock, runtime_names, exceptions, date.today())
    for message in accepted:
        print('REVIEWED:', message)
    for message in failures:
        print('BLOCKED:', message)
    print(f'npm: {len(failures)} blocking advisories; {len(accepted)} time-limited build exceptions.')
    return bool(failures)


if __name__ == '__main__':
    raise SystemExit(main())
