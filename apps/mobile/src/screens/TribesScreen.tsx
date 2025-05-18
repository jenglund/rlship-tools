import React, { useState, useEffect } from 'react';
import { View, StyleSheet, FlatList } from 'react-native';
import { Card, Text, ActivityIndicator, useTheme } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { RootStackParamList } from '../types/navigation';
import { useAuth } from '../contexts/AuthContext';

// Mock tribe data
interface Tribe {
  id: string;
  name: string;
  description: string;
  memberCount: number;
}

// Mock API call to get tribes
const fetchUserTribes = async (userId: string): Promise<Tribe[]> => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 500));
  
  // Mock data
  return [
    {
      id: 'tribe1',
      name: 'Family',
      description: 'Family activities and gatherings',
      memberCount: 4,
    },
    {
      id: 'tribe2',
      name: 'Friends',
      description: 'Weekend adventures with friends',
      memberCount: 6,
    },
    {
      id: 'tribe3',
      name: 'Work Buddies',
      description: 'Work team activities',
      memberCount: 8,
    },
    {
      id: 'tribe4',
      name: 'Book Club',
      description: 'Monthly book discussions',
      memberCount: 5,
    },
  ];
};

export default function TribesScreen() {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const theme = useTheme();
  const { user } = useAuth();
  
  const [tribes, setTribes] = useState<Tribe[]>([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    const loadTribes = async () => {
      if (user) {
        try {
          const userTribes = await fetchUserTribes(user.id);
          setTribes(userTribes);
        } catch (error) {
          console.error('Error loading tribes:', error);
          // In a real app, we'd show an error message here
        } finally {
          setLoading(false);
        }
      }
    };
    
    loadTribes();
  }, [user]);
  
  const handleTribePress = (tribeId: string) => {
    navigation.navigate('TribeDetails', { tribeId });
  };
  
  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={theme.colors.primary} />
        <Text style={styles.loadingText}>Loading your tribes...</Text>
      </View>
    );
  }
  
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Your Tribes</Text>
      
      {tribes.length === 0 ? (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>You aren't a member of any tribes yet.</Text>
        </View>
      ) : (
        <FlatList
          data={tribes}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => (
            <Card 
              style={styles.tribeCard} 
              onPress={() => handleTribePress(item.id)}
            >
              <Card.Content>
                <Text style={styles.tribeName}>{item.name}</Text>
                <Text style={styles.tribeDescription}>{item.description}</Text>
                <Text style={styles.memberCount}>{item.memberCount} members</Text>
              </Card.Content>
            </Card>
          )}
          contentContainerStyle={styles.tribesList}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  tribesList: {
    paddingBottom: 16,
  },
  tribeCard: {
    marginBottom: 12,
  },
  tribeName: {
    fontSize: 18,
    fontWeight: 'bold',
  },
  tribeDescription: {
    marginTop: 4,
    marginBottom: 8,
  },
  memberCount: {
    fontSize: 12,
    opacity: 0.7,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
    textAlign: 'center',
    opacity: 0.7,
  },
}); 