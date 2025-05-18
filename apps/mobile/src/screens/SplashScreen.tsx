import React, { useEffect } from 'react';
import { View, StyleSheet } from 'react-native';
import { Button, Text } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { RootStackParamList } from '../types/navigation';
import { useAuth } from '../contexts/AuthContext';

/**
 * SplashScreen component that displays the Tribe logo and a login button
 * Automatically redirects to MainTabs if already authenticated
 */
export default function SplashScreen() {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    // Check if user is already authenticated and redirect after a short delay
    if (!isLoading && isAuthenticated) {
      const timer = setTimeout(() => {
        navigation.reset({
          index: 0,
          routes: [{ name: 'MainTabs' }],
        });
      }, 1000); // Short delay to show the splash screen

      return () => clearTimeout(timer);
    }
  }, [isLoading, isAuthenticated, navigation]);

  const handleLogin = () => {
    navigation.navigate('Login');
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Tribe</Text>
      <Button
        mode="contained"
        onPress={handleLogin}
        style={styles.button}
        labelStyle={styles.buttonText}
      >
        Log In
      </Button>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  title: {
    fontSize: 48,
    fontWeight: 'bold',
    marginBottom: 48,
  },
  button: {
    width: '80%',
    paddingVertical: 8,
  },
  buttonText: {
    fontSize: 18,
  },
}); 