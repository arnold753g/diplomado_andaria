// Regenerate the web and Android icons from the supplied brand artwork.
// Requires sharp (available in the Codex workspace runtime), via NODE_PATH.
const fs = require('node:fs/promises')
const path = require('node:path')
const sharp = require('sharp')

const root = path.resolve(__dirname, '../..')
const source = path.join(root, 'assets/branding/andaria-icon-source.png')
const output = (name) => path.join(root, name)

async function write(name, bytes) {
  await fs.mkdir(path.dirname(output(name)), { recursive: true })
  await fs.writeFile(output(name), bytes)
}

async function main() {
  const { data, info } = await sharp(source).ensureAlpha().raw().toBuffer({ resolveWithObject: true })
  let left = info.width, top = info.height, right = 0, bottom = 0
  for (let y = 0; y < info.height; y++) {
    for (let x = 0; x < info.width; x++) {
      if (data[(y * info.width + x) * info.channels + 3]) {
        left = Math.min(left, x); top = Math.min(top, y)
        right = Math.max(right, x + 1); bottom = Math.max(bottom, y + 1)
      }
    }
  }
  if (right <= left || bottom <= top) throw new Error('The source artwork is empty')
  const artwork = await sharp(source).extract({ left, top, width: right - left, height: bottom - top }).png().toBuffer()
  async function icon(size, coverage) {
    const edge = Math.round(size * coverage)
    const mark = await sharp(artwork).resize(edge, edge, { fit: 'inside' }).png().toBuffer()
    return sharp({ create: { width: size, height: size, channels: 4, background: '#00000000' } })
      .composite([{ input: mark, gravity: 'centre' }]).png().toBuffer()
  }
  const large = await icon(512, 0.88)
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><image width="512" height="512" href="data:image/png;base64,${large.toString('base64')}"/></svg>\n`
  await write('frontend/public/branding/favicon.svg', svg)
  await write('mobile/web/favicon.svg', svg)
  await write('frontend/public/branding/apple-touch-icon.png', await icon(180, 0.8))
  await write('frontend/public/branding/andaria-icon.png', large)
  await write('mobile/assets/branding/andaria-symbol.png', large)
  await write('mobile/web/icons/Icon-192.png', await icon(192, 0.8))
  await write('mobile/web/icons/Icon-512.png', await icon(512, 0.8))
  await write('mobile/web/apple-touch-icon.png', await icon(180, 0.8))

  const sizes = [16, 32, 48]
  const images = await Promise.all(sizes.map(size => icon(size, 0.88)))
  const header = Buffer.alloc(6 + sizes.length * 16)
  header.writeUInt16LE(1, 2); header.writeUInt16LE(sizes.length, 4)
  let offset = header.length
  sizes.forEach((size, i) => {
    const entry = 6 + i * 16
    header[entry] = size; header[entry + 1] = size
    header.writeUInt16LE(1, entry + 4); header.writeUInt16LE(32, entry + 6)
    header.writeUInt32LE(images[i].length, entry + 8); header.writeUInt32LE(offset, entry + 12)
    offset += images[i].length
  })
  const ico = Buffer.concat([header, ...images])
  await write('frontend/public/favicon.ico', ico)
  await write('mobile/web/favicon.ico', ico)

  for (const [density, size] of Object.entries({ mdpi: 48, hdpi: 72, xhdpi: 96, xxhdpi: 144, xxxhdpi: 192 })) {
    await write(`mobile/android/app/src/main/res/mipmap-${density}/ic_launcher.png`, await icon(size, 0.64))
  }
  // 108dp canvas at xxxhdpi, with the entire square inside the 66dp safe circle.
  await write('mobile/android/app/src/main/res/drawable-xxxhdpi/andaria_icon_foreground.png', await icon(432, 46 / 108))
  console.log('Andaria icons generated from the supplied artwork; transparency, colours and aspect ratio preserved.')
}

main().catch(error => { console.error(error); process.exitCode = 1 })
