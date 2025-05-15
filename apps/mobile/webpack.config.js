const createExpoWebpackConfigAsync = require('@expo/webpack-config');
const path = require('path');

module.exports = async function (env, argv) {
  const config = await createExpoWebpackConfigAsync(
    {
      ...env,
      babel: {
        dangerouslyAddModulePathsToTranspile: [
          '@ui-kitten/components',
          'react-native-vector-icons'
        ]
      }
    },
    argv
  );
  
  // Customize the config before returning it.
  config.resolve.alias = {
    ...config.resolve.alias,
    'react-native-vector-icons': '@expo/vector-icons',
  };

  // Set the entry point
  config.entry = {
    app: [
      path.resolve(__dirname, 'index.web.js')
    ]
  };

  // Add fallbacks for node core modules
  config.resolve.fallback = {
    ...config.resolve.fallback,
    crypto: require.resolve('crypto-browserify'),
    stream: require.resolve('stream-browserify'),
    buffer: require.resolve('buffer/'),
  };

  return config;
}; 