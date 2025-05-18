import React, { useEffect, useState } from 'react';
import { View, StyleSheet, Animated, Dimensions, TouchableOpacity } from 'react-native';
import { Text } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { RootStackParamList } from '../types/navigation';
import { useAuth } from '../contexts/AuthContext';
import { MaterialCommunityIcons } from '@expo/vector-icons';

/**
 * SplashScreen component that displays the Tribe logo and a login button
 * Automatically redirects to MainTabs if already authenticated
 */
export default function SplashScreen() {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { isAuthenticated, isLoading } = useAuth();
  const [fadeAnim] = useState(new Animated.Value(0));
  const [scaleAnim] = useState(new Animated.Value(0.9));

  useEffect(() => {
    // Animation sequence
    Animated.parallel([
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 1000,
        useNativeDriver: true,
      }),
      Animated.spring(scaleAnim, {
        toValue: 1,
        friction: 8,
        useNativeDriver: true,
      })
    ]).start();

    // Check if user is already authenticated and redirect after animation completes
    if (!isLoading && isAuthenticated) {
      const timer = setTimeout(() => {
        navigation.reset({
          index: 0,
          routes: [{ name: 'MainTabs' }],
        });
      }, 1500);

      return () => clearTimeout(timer);
    }
  }, [isLoading, isAuthenticated, navigation, fadeAnim, scaleAnim]);

  const handleLogin = () => {
    navigation.navigate('Login');
  };

  const { width } = Dimensions.get('window');
  const iconSize = width * 0.4;

  return (
    <View style={styles.container}>
      <Animated.View style={[
        styles.logoContainer, 
        { 
          opacity: fadeAnim,
          transform: [{ scale: scaleAnim }] 
        }
      ]}>
        <MaterialCommunityIcons 
          name="account-group" 
          size={iconSize} 
          color="#6200ee" 
        />
        <Text style={styles.title}>TRIBE</Text>
        <Text style={styles.tagline}>Share experiences with your people</Text>
      </Animated.View>

      <Animated.View style={{ opacity: fadeAnim }}>
        <TouchableOpacity 
          style={styles.button}
          onPress={handleLogin}
          activeOpacity={0.8}
        >
          <Text style={styles.buttonText}>Get Started</Text>
        </TouchableOpacity>
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
    backgroundColor: '#f6f6f6',
  },
  logoContainer: {
    alignItems: 'center',
    marginBottom: 60,
  },
  title: {
    fontSize: 64,
    fontWeight: 'bold',
    marginTop: 16,
    color: '#6200ee',
    letterSpacing: 2,
  },
  tagline: {
    fontSize: 18,
    marginTop: 8,
    textAlign: 'center',
    color: '#666',
  },
  button: {
    width: 220,
    paddingVertical: 14,
    borderRadius: 30,
    backgroundColor: '#6200ee',
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 3.84,
    elevation: 5,
  },
  buttonText: {
    fontSize: 20,
    color: '#fff',
    fontWeight: 'bold',
  },
}); 