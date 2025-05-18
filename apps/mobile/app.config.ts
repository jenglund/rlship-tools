import { ExpoConfig } from '@expo/config-types';

export default (): ExpoConfig => ({
  name: 'RLShip Tools',
  slug: 'rlship-tools',
  version: '1.0.0',
  orientation: 'portrait',
  icon: './assets/icon.png',
  userInterfaceStyle: 'light',
  splash: {
    image: './assets/splash.png',
    resizeMode: 'contain',
    backgroundColor: '#ffffff'
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
    name: 'RLShip Tools',
    shortName: 'RLShip',
    lang: 'en',
    themeColor: '#ffffff',
    backgroundColor: '#ffffff',
    description: 'Relationship management tools for couples and polyamorous relationships',
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