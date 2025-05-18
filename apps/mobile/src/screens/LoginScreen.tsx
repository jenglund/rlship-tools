import React from 'react';
import { View, StyleSheet } from 'react-native';
import { Button, Text, useTheme } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { RootStackParamList } from '../types/navigation';
import { useAuth } from '../contexts/AuthContext';

/**
 * LoginScreen component that displays buttons to log in as various dev users
 */
export default function LoginScreen() {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const theme = useTheme();
  const { login } = useAuth();
  
  const handleLogin = async (email: string) => {
    try {
      await login(email);
      navigation.reset({
        index: 0,
        routes: [{ name: 'MainTabs' }],
      });
    } catch (error) {
      console.error('Login failed:', error);
      // In a real app, we'd show an error toast/alert here
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Log In</Text>
      <Text style={styles.subtitle}>Select a development user:</Text>
      
      <View style={styles.buttonContainer}>
        <Button
          mode="contained"
          onPress={() => handleLogin('dev_user1@gonetribal.com')}
          style={styles.button}
          labelStyle={styles.buttonText}
        >
          Log in as dev_user1
        </Button>
        
        <Button
          mode="contained"
          onPress={() => handleLogin('dev_user2@gonetribal.com')}
          style={styles.button}
          labelStyle={styles.buttonText}
        >
          Log in as dev_user2
        </Button>
        
        <Button
          mode="contained"
          onPress={() => handleLogin('dev_user3@gonetribal.com')}
          style={styles.button}
          labelStyle={styles.buttonText}
        >
          Log in as dev_user3
        </Button>
        
        <Button
          mode="contained"
          onPress={() => handleLogin('dev_user4@gonetribal.com')}
          style={styles.button}
          labelStyle={styles.buttonText}
        >
          Log in as dev_user4
        </Button>
      </View>
      
      {/* This will be enabled in the future */}
      <Button
        mode="outlined"
        disabled
        style={styles.googleButton}
      >
        Log in with Google
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
    fontSize: 32,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  subtitle: {
    fontSize: 16,
    marginBottom: 24,
    textAlign: 'center',
  },
  buttonContainer: {
    width: '100%',
    marginBottom: 24,
  },
  button: {
    marginBottom: 12,
    width: '100%',
  },
  buttonText: {
    fontSize: 16,
  },
  googleButton: {
    width: '100%',
  },
}); 