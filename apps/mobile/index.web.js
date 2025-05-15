import { registerRootComponent } from 'expo';
import { Platform } from 'react-native';
import App from './App';

// Register the app component for web
if (Platform.OS === 'web') {
  const rootTag = document.getElementById('root');
  registerRootComponent(App, { rootTag });
} else {
  registerRootComponent(App);
} 