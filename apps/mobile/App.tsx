import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { Provider as PaperProvider } from 'react-native-paper';
import { StatusBar } from 'expo-status-bar';

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

const Tab = createBottomTabNavigator<TabParamList>();
const Stack = createNativeStackNavigator<RootStackParamList>();
const TribesStack = createNativeStackNavigator<TribesStackParamList>();

// Tribes stack navigator component
function TribesStackNavigator() {
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
  return (
    <Tab.Navigator>
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
  return (
    <PaperProvider>
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
  );
}
