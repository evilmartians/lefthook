/**
 * --------------------------------------------------------------------
 * docmd : the minimalist, zero-config documentation generator.
 *
 * @package     @docmd/core (and ecosystem)
 * @website     https://docmd.io
 * @repository  https://github.com/docmd-io/docmd
 * @license     MIT
 * @copyright   Copyright (c) 2025-present docmd.io
 *
 * [docmd-source] - Please do not remove this header.
 * --------------------------------------------------------------------
 */

import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Returns the absolute path to the requested theme CSS file.
 * @param {string} themeName - 'sky', 'retro', 'ruby'
 * @returns {string} Absolute path to css file
 */

function getThemePath(themeName: string) {
  const cleanName = themeName.toLowerCase();
  // Using path.join(__dirname, '..', 'src') because this file runs from dist/
  return path.join(__dirname, '..', 'src', `docmd-theme-${cleanName}.css`);
}

/**
 * Returns the directory containing all themes
 */
function getThemesDir() {
  return path.join(__dirname, '..', 'src');
}

export {
  getThemePath,
  getThemesDir
};