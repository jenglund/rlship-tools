import React, { useState, useEffect } from 'react';
import { View, StyleSheet, FlatList } from 'react-native';
import { Text, Avatar, Divider, ActivityIndicator, useTheme } from 'react-native-paper';
import { useRoute, useNavigation, RouteProp } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { RootStackParamList } from '../types/navigation';

// Mock tribe member data
interface TribeMember {
  id: string;
  displayName: string;
  email: string;
  avatarUrl?: string;
}

// Mock tribe data
interface TribeDetails {
  id: string;
  name: string;
  description: string;
  members: TribeMember[];
}

// Mock API call to get tribe details
const fetchTribeDetails = async (tribeId: string): Promise<TribeDetails> => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 600));
  
  // Mock data based on tribe ID
  const mockTribes: Record<string, TribeDetails> = {
    'tribe1': {
      id: 'tribe1',
      name: 'Family',
      description: 'Family activities and gatherings',
      members: [
        { id: 'member1', displayName: 'John Doe', email: 'john@example.com' },
        { id: 'member2', displayName: 'Jane Doe', email: 'jane@example.com' },
        { id: 'member3', displayName: 'Kid Doe', email: 'kid@example.com' },
        { id: 'member4', displayName: 'Baby Doe', email: 'baby@example.com' },
      ],
    },
    'tribe2': {
      id: 'tribe2',
      name: 'Friends',
      description: 'Weekend adventures with friends',
      members: [
        { id: 'member1', displayName: 'Alice Smith', email: 'alice@example.com' },
        { id: 'member2', displayName: 'Bob Johnson', email: 'bob@example.com' },
        { id: 'member3', displayName: 'Carol Williams', email: 'carol@example.com' },
        { id: 'member4', displayName: 'Dave Brown', email: 'dave@example.com' },
        { id: 'member5', displayName: 'Eve Davis', email: 'eve@example.com' },
        { id: 'member6', displayName: 'Frank Miller', email: 'frank@example.com' },
      ],
    },
    'tribe3': {
      id: 'tribe3',
      name: 'Work Buddies',
      description: 'Work team activities',
      members: [
        { id: 'member1', displayName: 'Manager', email: 'manager@example.com' },
        { id: 'member2', displayName: 'Developer 1', email: 'dev1@example.com' },
        { id: 'member3', displayName: 'Developer 2', email: 'dev2@example.com' },
        { id: 'member4', displayName: 'Designer', email: 'designer@example.com' },
        { id: 'member5', displayName: 'QA Engineer', email: 'qa@example.com' },
        { id: 'member6', displayName: 'Product Owner', email: 'po@example.com' },
        { id: 'member7', displayName: 'Scrum Master', email: 'scrum@example.com' },
        { id: 'member8', displayName: 'DevOps', email: 'devops@example.com' },
      ],
    },
    'tribe4': {
      id: 'tribe4',
      name: 'Book Club',
      description: 'Monthly book discussions',
      members: [
        { id: 'member1', displayName: 'George Wilson', email: 'george@example.com' },
        { id: 'member2', displayName: 'Helen Moore', email: 'helen@example.com' },
        { id: 'member3', displayName: 'Ivan Taylor', email: 'ivan@example.com' },
        { id: 'member4', displayName: 'Julia Anderson', email: 'julia@example.com' },
        { id: 'member5', displayName: 'Kevin Thomas', email: 'kevin@example.com' },
      ],
    },
  };
  
  const tribe = mockTribes[tribeId];
  if (!tribe) {
    throw new Error(`Tribe with ID ${tribeId} not found`);
  }
  
  return tribe;
};

export default function TribeDetailsScreen() {
  const route = useRoute<RouteProp<RootStackParamList, 'TribeDetails'>>();
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const theme = useTheme();
  const { tribeId } = route.params;
  
  const [tribe, setTribe] = useState<TribeDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const loadTribeDetails = async () => {
      try {
        const details = await fetchTribeDetails(tribeId);
        setTribe(details);
        navigation.setOptions({ title: details.name });
      } catch (err) {
        console.error('Error fetching tribe details:', err);
        setError('Could not load tribe details');
      } finally {
        setLoading(false);
      }
    };
    
    loadTribeDetails();
  }, [tribeId, navigation]);
  
  if (loading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={theme.colors.primary} />
        <Text style={styles.loadingText}>Loading tribe details...</Text>
      </View>
    );
  }
  
  if (error || !tribe) {
    return (
      <View style={styles.errorContainer}>
        <Text style={styles.errorText}>{error || 'An unknown error occurred'}</Text>
      </View>
    );
  }
  
  return (
    <View style={styles.container}>
      <View style={styles.headerContainer}>
        <Text style={styles.tribeName}>{tribe.name}</Text>
        <Text style={styles.tribeDescription}>{tribe.description}</Text>
        <Text style={styles.memberCountText}>{tribe.members.length} members</Text>
      </View>
      
      <Divider style={styles.divider} />
      
      <Text style={styles.sectionTitle}>Members</Text>
      
      <FlatList
        data={tribe.members}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <View style={styles.memberItem}>
            <Avatar.Text 
              size={40} 
              label={item.displayName.substring(0, 2)}
              color={theme.colors.onPrimary}
              style={{ backgroundColor: theme.colors.primary }}
            />
            <View style={styles.memberInfo}>
              <Text style={styles.memberName}>{item.displayName}</Text>
              <Text style={styles.memberEmail}>{item.email}</Text>
            </View>
          </View>
        )}
        contentContainerStyle={styles.membersList}
      />
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
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 16,
  },
  errorText: {
    color: 'red',
    textAlign: 'center',
    fontSize: 16,
  },
  headerContainer: {
    marginBottom: 16,
  },
  tribeName: {
    fontSize: 24,
    fontWeight: 'bold',
  },
  tribeDescription: {
    fontSize: 16,
    marginTop: 8,
    marginBottom: 8,
  },
  memberCountText: {
    fontSize: 14,
    opacity: 0.7,
  },
  divider: {
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  membersList: {
    paddingBottom: 16,
  },
  memberItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  memberInfo: {
    marginLeft: 12,
  },
  memberName: {
    fontSize: 16,
    fontWeight: 'bold',
  },
  memberEmail: {
    fontSize: 14,
    opacity: 0.7,
  },
}); 