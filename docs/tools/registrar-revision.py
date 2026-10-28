"""Registro local de revisiones: begin, close y render. Sólo fuentes públicas."""
from pathlib import Path
from datetime import datetime
import argparse, difflib, hashlib, html, json, os

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / 'docs/revisiones'
MONO = ROOT / 'documentacion para el diplomando/Documento del sistema Texstudio/Monografia_UAJMS_Diplomado_Web_Movil'

def snapshot(scope_version=2):
    candidates = [ROOT/'README.md', ROOT/'AGENTS.md', ROOT/'nginx/default.conf']
    candidates += list(ROOT.glob('docker-compose*.yml'))
    # Podar dependencias y archivos generados antes de recorrerlos.
    # Evita que cada captura explore node_modules o las cachés Android.
    excluded={'node_modules','vendor','build','.nuxt','.output','.dart_tool','.gradle','.git','__pycache__','revisiones','evidencias'}
    folders=[ROOT/'backend',ROOT/'docs',MONO]
    if scope_version>=2:folders += [ROOT/'frontend',ROOT/'mobile']
    for folder in folders:
        for parent,dirs,names in os.walk(folder):
            dirs[:]=[name for name in dirs if name not in excluded]
            for name in names:
                path=Path(parent)/name;local=path.relative_to(folder).as_posix()
                keep=False
                if folder==ROOT/'backend':
                    keep=path.suffix=='.go' or local in ['go.mod','go.sum'] or (scope_version>=2 and (local=='Dockerfile' or local.startswith('migrations/') and path.suffix=='.sql'))
                elif folder==ROOT/'docs':
                    keep=path.suffix=='.md' or path.parent==folder/'tools' and path.suffix in ['.py','.cjs'] or path.parent==folder/'api' and path.suffix=='.json'
                elif folder==MONO:keep=path.suffix in ['.tex','.bib']
                elif folder==ROOT/'frontend':keep=path.suffix in ['.vue','.ts','.js','.css','.scss'] or local in ['package.json','tsconfig.json','Dockerfile']
                elif folder==ROOT/'mobile':
                    first=local.split('/')[0]
                    keep=(first in ['lib','test','integration_test','test_driver'] and path.suffix=='.dart') or local in ['pubspec.yaml','analysis_options.yaml'] or (first=='android' and (path.suffix in ['.xml','.gradle','.kt','.java'] or name.endswith('.gradle.kts'))) or (local.startswith('ios/Runner/') and path.suffix in ['.swift','.plist'])
                if keep:candidates.append(path)
    result = {}
    for path in sorted(set(candidates)):
        relative = path.relative_to(ROOT).as_posix()
        if not path.is_file() or any(x in relative.lower() for x in ['/revisiones/','/evidencias/','/node_modules/','/vendor/','/build/','/.nuxt/','/.output/','/.dart_tool/','privado','acceso-produccion']): continue
        result[relative] = path.read_text(encoding='utf-8-sig')
    return result

def save(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')

def render():
    sections = []
    for folder in sorted(OUT.glob('[0-9]*'), reverse=True):
        meta = json.loads((folder/'revision.json').read_text(encoding='utf-8'))
        before = json.loads((folder/'antes.json').read_text(encoding='utf-8'))
        after = json.loads((folder/'despues.json').read_text(encoding='utf-8')) if (folder/'despues.json').exists() else snapshot(meta.get('scope_version',1))
        changes = []
        for name in sorted(before.keys() | after.keys()):
            old, new = before.get(name,''), after.get(name,'')
            if old == new: continue
            rows = []
            a, b = old.splitlines(), new.splitlines()
            matcher = difflib.SequenceMatcher(None,a,b,autojunk=False)
            for group in matcher.get_grouped_opcodes(2):
                rows.append('<tr><td colspan="2" class="context">Fragmento modificado</td></tr>')
                for tag,i,j,k,l in group:
                    for n in range(max(j-i,l-k)):
                        left = html.escape(b[k+n]) if k+n<l else ''
                        right = html.escape(a[i+n]) if i+n<j else ''
                        css = 'same' if tag=='equal' else 'change'
                        rows.append(f'<tr class="{css}"><td><span class="number">{k+n+1 if k+n<l else ""}</span><code>{left}</code></td><td><span class="number">{i+n+1 if i+n<j else ""}</span><code>{right}</code></td></tr>')
            changes.append(f'<details class="file"><summary>{html.escape(name)}</summary><div class="scroll"><table><thead><tr><th>MODIFICADO · nuevo</th><th>ANTERIOR · retirado</th></tr></thead><tbody>{"".join(rows)}</tbody></table></div></details>')
        state = 'Cerrada' if meta.get('closed_at') else 'En curso'
        sections.append(f'<section><h2>{html.escape(meta["id"])} · {html.escape(meta["title"])}</h2><p>{state} · inicio {html.escape(meta["started_at"])} · {len(changes)} archivos con cambios</p><p>{html.escape(meta.get("result",""))}</p>{"".join(changes) or "<p>Sin cambios de fuentes todavía.</p>"}</section>')
    content = '''<!doctype html><html lang="es"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Andaria · Registro de revisiones</title><style>
    *{box-sizing:border-box}body{margin:0;background:#f3f5f8;color:#172236;font:16px system-ui,sans-serif}header,main{max-width:1500px;margin:auto;padding:24px}header{background:#172b49;color:white;max-width:none}header>div{max-width:1450px;margin:auto}h1{margin:0 0 10px;font-size:28px}p{line-height:1.5}section{background:white;border:1px solid #d7dfe9;border-radius:12px;padding:22px;margin:0 0 24px}h2{font-size:21px}.file{margin:12px 0;border:1px solid #bac6d6;border-radius:6px;overflow:hidden}summary{cursor:pointer;padding:12px;background:#eef2f7;overflow-wrap:anywhere;font-weight:600}.scroll{overflow:auto}table{border-collapse:collapse;table-layout:fixed;width:100%;min-width:850px}th{padding:12px;text-align:left;width:50%;background:#d9f4df}th+th{background:#ffe0e0}td{vertical-align:top;padding:4px 8px;border-right:1px solid #d0d9e3;overflow-wrap:anywhere}tr.change td{background:#e5f7e9;border-left:3px solid #228043}tr.change td+td{background:#ffeaea;border-left:3px solid #b53333}.context{background:#e6edf6;font-size:12px;font-weight:bold}.same{color:#627087}.number{display:inline-block;width:42px;color:#687681;font:12px monospace;user-select:none}code{font:13px/1.45 Consolas,monospace;white-space:pre-wrap}input,button{font:inherit;border-radius:6px;padding:9px;border:1px solid #a5b4c7;margin:5px}input{width:min(600px,95%)}.legend{display:flex;gap:18px;flex-wrap:wrap}.legend span{padding:7px 12px;border-radius:5px;color:#142b20;background:#d9f4df}.legend span+span{background:#ffe0e0;color:#502323}@media(max-width:600px){main{padding:10px}section{padding:12px}}
    </style><header><div><h1>Andaria · Registro de revisiones</h1><p>Fuentes comparadas antes y después de cada revisión. El registro comienza el 11/10/2026; no reconstruye cambios anteriores.</p><div class="legend"><span>Izquierda: modificado / nuevo</span><span>Derecha: anterior / retirado</span></div></div></header><main><label>Buscar archivo o revisión <input id="filter" type="search" placeholder="Ej.: capitulo2, salud, antecedentes"></label><button id="expand">Abrir todos</button><button id="collapse">Cerrar todos</button>'''+''.join(sections)+'''</main><script>
    const files=[...document.querySelectorAll('.file')];document.querySelector('#filter').oninput=e=>{const q=e.target.value.toLowerCase();files.forEach(f=>f.hidden=!f.textContent.toLowerCase().includes(q));document.querySelectorAll('section').forEach(s=>s.hidden=q&&!s.textContent.toLowerCase().includes(q))};document.querySelector('#expand').onclick=()=>files.filter(f=>!f.hidden).forEach(f=>f.open=true);document.querySelector('#collapse').onclick=()=>files.forEach(f=>f.open=false);
    </script></html>'''
    (OUT/'index.html').write_text(content,encoding='utf-8')

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('action',choices=['begin','close','render'])
p.add_argument('--title',default='')
p.add_argument('--result',default='')
args=p.parse_args()
OUT.mkdir(parents=True,exist_ok=True)
active=OUT/'activa.json'
if args.action=='begin':
    if active.exists(): raise SystemExit('Hay una revisión abierta. Ciérrela antes de comenzar otra.')
    if not args.title: raise SystemExit('Se requiere --title.')
    number=len(list(OUT.glob('[0-9]*')))+1
    ident=f'{number:03d}-{datetime.now():%Y-%m-%d}'
    meta={'id':ident,'title':args.title,'started_at':datetime.now().astimezone().isoformat(),'scope_version':2}
    save(OUT/ident/'antes.json',snapshot());save(OUT/ident/'revision.json',meta);save(active,{'id':ident})
elif args.action=='close':
    ident=json.loads(active.read_text(encoding='utf-8'))['id']
    path=OUT/ident/'revision.json';meta=json.loads(path.read_text(encoding='utf-8'))
    save(OUT/ident/'despues.json',snapshot(meta.get('scope_version',1)))
    meta.update(closed_at=datetime.now().astimezone().isoformat(),result=args.result)
    save(path,meta);active.unlink()
render()
print('Registro: '+str(OUT/'index.html'))
