// Learn more https://docs.expo.io/guides/customizing-metro
const { getDefaultConfig } = require('@expo/metro-config');

const defaultConfig = getDefaultConfig(__dirname);

// Expo 50 compatible configuration
defaultConfig.resolver.sourceExts = ['jsx', 'js', 'ts', 'tsx', 'json', 'web.js', 'web.jsx', 'web.ts', 'web.tsx'];
defaultConfig.resolver.assetExts = [...defaultConfig.resolver.assetExts, 'db', 'mp3', 'ttf', 'obj', 'png', 'jpg'];

module.exports = defaultConfig; 