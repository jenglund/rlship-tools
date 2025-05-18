import React from 'react';
import { View, StyleSheet } from 'react-native';
import { Text, Card, Button } from 'react-native-paper';

export default function HomeScreen() {
  return (
    <View style={styles.container}>
      <Card style={styles.card}>
        <Card.Title title="Recent Activities" />
        <Card.Content>
          <Text variant="bodyMedium">
            View your recent activities and memories together.
          </Text>
        </Card.Content>
        <Card.Actions>
          <Button>View All</Button>
        </Card.Actions>
      </Card>

      <Card style={styles.card}>
        <Card.Title title="Active Interest Buttons" />
        <Card.Content>
          <Text variant="bodyMedium">
            See what activities your partner(s) are interested in right now.
          </Text>
        </Card.Content>
        <Card.Actions>
          <Button>View Interests</Button>
        </Card.Actions>
      </Card>

      <Card style={styles.card}>
        <Card.Title title="Upcoming Plans" />
        <Card.Content>
          <Text variant="bodyMedium">
            Check out your upcoming planned activities.
          </Text>
        </Card.Content>
        <Card.Actions>
          <Button>View Plans</Button>
        </Card.Actions>
      </Card>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#f5f5f5',
  },
  card: {
    marginBottom: 16,
  },
}); 