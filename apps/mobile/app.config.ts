import { ExpoConfig, ConfigContext } from 'expo/config';

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
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
    favicon: './assets/favicon.png'
  },
  extra: {
    apiUrl: process.env.API_URL || 'http://localhost:8080',
  },
  plugins: [
    'expo-image-picker',
    'expo-notifications'
  ]
}); 