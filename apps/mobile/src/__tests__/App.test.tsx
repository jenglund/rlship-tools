import React from 'react';
import renderer from 'react-test-renderer';
import App from '../../App';

// Mock any native modules or components that might cause issues
jest.mock('react-native/Libraries/Animated/NativeAnimatedHelper');

describe('App', () => {
  it('renders without crashing', () => {
    expect(() => renderer.create(<App />)).not.toThrow();
  });
}); 