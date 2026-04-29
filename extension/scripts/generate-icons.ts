import sharp from 'sharp';
import { existsSync, mkdirSync } from 'fs';
import { join } from 'path';

const sizes = [16, 32, 48, 96, 128];
const inputSvg = join(process.cwd(), 'assets/icon.svg');
const outputDir = join(process.cwd(), 'public/icon');

if (!existsSync(outputDir)) {
  mkdirSync(outputDir, { recursive: true });
}

async function generateIcons() {
  console.log('Generating icons from SVG...');
  
  for (const size of sizes) {
    const outputFile = join(outputDir, `${size}.png`);
    await sharp(inputSvg)
      .resize(size, size)
      .png()
      .toFile(outputFile);
    console.log(`  Generated ${size}x${size} icon`);
  }
  
  console.log('Done!');
}

generateIcons().catch(console.error);
