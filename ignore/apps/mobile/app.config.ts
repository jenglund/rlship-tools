import { ExpoConfig } from '@expo/config-types';

export default (): ExpoConfig => ({
  name: 'Tribe',
  slug: 'rlship-tools',
  version: '1.0.0',
  orientation: 'portrait',
  icon: './assets/icon.png',
  userInterfaceStyle: 'light',
  splash: {
    image: './assets/splash-icon.png',
    resizeMode: 'contain',
    backgroundColor: '#f6f6f6'
  },
  assetBundlePatterns: [
    '**/*'
  ],
  ios: {
    supportsTablet: true,
    bundleIdentifier: 'com.yourdomain.rlshiptools'
  },
  android: {
    adaptiveIcon: {
      foregroundImage: './assets/adaptive-icon.png',
      backgroundColor: '#ffffff'
    },
    package: 'com.yourdomain.rlshiptools'
  },
  web: {
    favicon: './assets/favicon.png',
    name: 'Tribe',
    shortName: 'Tribe',
    lang: 'en',
    themeColor: '#6200ee',
    backgroundColor: '#f6f6f6',
    description: 'Share experiences with your tribes',
    orientation: 'any',
    scope: '/',
    startUrl: '/',
    display: 'standalone',
    crossorigin: 'use-credentials'
  },
  extra: {
    apiUrl: process.env.API_URL || 'http://localhost:8080',
  }
}); 