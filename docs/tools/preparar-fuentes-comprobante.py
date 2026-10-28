"""Genera instancias estaticas de Outfit para el motor PDF (solo desarrollo)."""
from pathlib import Path
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont

root = Path(__file__).resolve().parents[2]
source = root / 'mobile/assets/fonts/Outfit.ttf'
output = root / 'backend/internal/receipts/assets'
for name, weight in [('Regular', 400), ('SemiBold', 600)]:
    font = TTFont(source)
    instantiateVariableFont(font, {'wght': weight}, inplace=True)
    font.save(output / f'Outfit-{name}.ttf')
    font.close()
