// @ts-ignore - Fix for TypeScript errors with React Navigation
import React, { useEffect, useState } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { Provider as PaperProvider, DefaultTheme } from 'react-native-paper';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { View, Text, StyleSheet, ActivityIndicator } from 'react-native';
import { MaterialCommunityIcons } from '@expo/vector-icons';

// Import screens
import HomeScreen from './src/screens/HomeScreen';
import ListsScreen from './src/screens/ListsScreen';
import InterestButtonsScreen from './src/screens/InterestButtonsScreen';
import ProfileScreen from './src/screens/ProfileScreen';
import SplashScreen from './src/screens/SplashScreen';
import LoginScreen from './src/screens/LoginScreen';
import TribesScreen from './src/screens/TribesScreen';
import TribeDetailsScreen from './src/screens/TribeDetailsScreen';

// Import context providers
import { AuthProvider } from './src/contexts/AuthContext';

// Import types
import { RootStackParamList, TabParamList, TribesStackParamList } from './src/types/navigation';

// Define theme
const theme = {
  ...DefaultTheme,
  colors: {
    ...DefaultTheme.colors,
    primary: '#6200ee',
    accent: '#03dac4',
    background: '#f6f6f6',
  },
};

const Tab = createBottomTabNavigator<TabParamList>();
const Stack = createNativeStackNavigator<RootStackParamList>();
const TribesStack = createNativeStackNavigator<TribesStackParamList>();

// Define icon mapping
const ICON_NAMES: Record<string, keyof typeof MaterialCommunityIcons.glyphMap> = {
  Home: 'home',
  Tribes: 'account-group',
  Lists: 'format-list-bulleted',
  Interest: 'heart',
  Profile: 'account',
};

// Tribes stack navigator component
function TribesStackNavigator() {
  // @ts-ignore - Suppressing TypeScript errors with React Navigation
  return (
    <TribesStack.Navigator>
      <TribesStack.Screen 
        name="Tribes" 
        component={TribesScreen}
        options={{
          title: 'Your Tribes'
        }}
      />
      <TribesStack.Screen 
        name="TribeDetails" 
        component={TribeDetailsScreen}
        options={{
          title: 'Tribe Details'
        }}
      />
    </TribesStack.Navigator>
  );
}

// Tab navigator component
function TabNavigator() {
  // @ts-ignore - Suppressing TypeScript errors with React Navigation
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        tabBarIcon: ({ color, size }) => {
          const iconName = ICON_NAMES[route.name] || 'help-circle';
          return <MaterialCommunityIcons name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: theme.colors.primary,
        tabBarInactiveTintColor: 'gray',
      })}
    >
      <Tab.Screen 
        name="Home" 
        component={HomeScreen}
        options={{
          title: 'Home'
        }}
      />
      <Tab.Screen 
        name="Tribes" 
        component={TribesStackNavigator}
        options={{
          title: 'Tribes',
          headerShown: false
        }}
      />
      <Tab.Screen 
        name="Lists" 
        component={ListsScreen}
        options={{
          title: 'Lists'
        }}
      />
      <Tab.Screen 
        name="Interest" 
        component={InterestButtonsScreen}
        options={{
          title: 'Interest Buttons'
        }}
      />
      <Tab.Screen 
        name="Profile" 
        component={ProfileScreen}
        options={{
          title: 'Profile'
        }}
      />
    </Tab.Navigator>
  );
}

export default function App() {
  const [isReady, setIsReady] = useState(false);

  // Simulate app initialization
  useEffect(() => {
    // Give time for resources to load
    const timer = setTimeout(() => {
      setIsReady(true);
    }, 2000); // Longer delay to ensure splash screen is visible
    
    return () => clearTimeout(timer);
  }, []);

  if (!isReady) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#6200ee" />
        <Text style={styles.loadingText}>Loading Tribe...</Text>
      </View>
    );
  }

  // @ts-ignore - Suppressing TypeScript errors with React Navigation
  return (
    <SafeAreaProvider>
      <PaperProvider theme={theme}>
        <AuthProvider>
          <NavigationContainer>
            <Stack.Navigator initialRouteName="Splash" screenOptions={{ headerShown: false }}>
              <Stack.Screen name="Splash" component={SplashScreen} />
              <Stack.Screen name="Login" component={LoginScreen} />
              <Stack.Screen
                name="MainTabs"
                component={TabNavigator}
              />
            </Stack.Navigator>
            <StatusBar style="auto" />
          </NavigationContainer>
        </AuthProvider>
      </PaperProvider>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f6f6f6',
  },
  loadingText: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#6200ee',
    marginTop: 16,
  },
});
