"""Check resolved Pub and Android Maven versions with the official OSV API."""
from pathlib import Path
import argparse
import json
import urllib.request


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--pub-deps', type=Path, required=True)
    parser.add_argument('--android-bom', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    pub = json.loads(args.pub_deps.read_text(encoding='utf-8-sig'))
    bom = json.loads(args.android_bom.read_text(encoding='utf-8-sig'))
    packages = {('Pub', item['name'], item['version']) for item in pub['packages'] if item.get('source') == 'hosted'}
    pub_count = len(packages)
    native = {('Maven', item['group'] + ':' + item['name'], item['version']) for item in bom['components']}
    if not pub_count or not native:
        raise SystemExit('Both resolved dependency inventories must be nonempty.')
    packages = sorted(packages | native)
    findings = []
    for offset in range(0, len(packages), 100):
        batch = packages[offset:offset + 100]
        body = {'queries': [{'package': {'ecosystem': ecosystem, 'name': name}, 'version': version}
                            for ecosystem, name, version in batch]}
        request = urllib.request.Request('https://api.osv.dev/v1/querybatch',
                                         data=json.dumps(body).encode(), headers={'Content-Type': 'application/json'})
        with urllib.request.urlopen(request, timeout=60) as response:
            results = json.load(response)['results']
        if len(results) != len(batch):
            raise SystemExit('Incomplete OSV response; refusing to pass the scan.')
        for package, result in zip(batch, results):
            if result.get('vulns'):
                findings.append({'ecosystem': package[0], 'name': package[1], 'version': package[2],
                                 'advisories': [item['id'] for item in result['vulns']]})
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps({'pub_packages': pub_count, 'android_packages': len(native), 'findings': findings}, indent=2) + '\n', encoding='utf-8')
    print(f'OSV: {pub_count} hosted Pub packages; {len(native)} Android Maven packages; {len(findings)} affected packages.')
    return bool(findings)


if __name__ == '__main__':
    raise SystemExit(main())
