import React, { useState, useCallback, useEffect } from 'react';
import { GoogleMap, useJsApiLoader, Marker, InfoWindow } from '@react-google-maps/api';
import { Card, Spinner, Alert } from 'react-bootstrap';

const containerStyle = {
  width: '100%',
  height: '400px',
  borderRadius: '8px'
};

const defaultCenter = {
  lat: 37.7749, // San Francisco default
  lng: -122.4194
};

// Google Maps API key should be stored in environment variable
const apiKey = process.env.REACT_APP_GOOGLE_MAPS_API_KEY || '';

const LocationMapView = ({ items = [], onItemClick = null }) => {
  const [map, setMap] = useState(null);
  const [selectedItem, setSelectedItem] = useState(null);
  const [center, setCenter] = useState(defaultCenter);
  const [bounds, setBounds] = useState(null);

  const { isLoaded, loadError } = useJsApiLoader({
    id: 'google-map-script',
    googleMapsApiKey: apiKey
  });

  // Reset bounds when items change
  useEffect(() => {
    if (isLoaded && items.length > 0 && map) {
      const newBounds = new window.google.maps.LatLngBounds();
      
      let hasValidLocations = false;
      
      // Add each item to the bounds if it has valid coordinates
      items.forEach(item => {
        if (item.metadata && item.metadata.latitude && item.metadata.longitude) {
          const lat = parseFloat(item.metadata.latitude);
          const lng = parseFloat(item.metadata.longitude);
          
          if (!isNaN(lat) && !isNaN(lng)) {
            newBounds.extend({ lat, lng });
            hasValidLocations = true;
          }
        }
      });
      
      // Only set bounds if we have valid locations
      if (hasValidLocations) {
        setBounds(newBounds);
        map.fitBounds(newBounds);
        
        // Set center to the first item that has coordinates
        for (const item of items) {
          if (item.metadata && item.metadata.latitude && item.metadata.longitude) {
            const lat = parseFloat(item.metadata.latitude);
            const lng = parseFloat(item.metadata.longitude);
            
            if (!isNaN(lat) && !isNaN(lng)) {
              setCenter({ lat, lng });
              break;
            }
          }
        }
      }
    }
  }, [isLoaded, items, map]);

  const onLoad = useCallback(map => {
    setMap(map);
  }, []);

  const onUnmount = useCallback(() => {
    setMap(null);
  }, []);

  const handleMarkerClick = (item) => {
    setSelectedItem(item);
    
    if (onItemClick) {
      onItemClick(item);
    }
  };

  const renderMarker = (item) => {
    if (!item.metadata || !item.metadata.latitude || !item.metadata.longitude) {
      return null;
    }
    
    const lat = parseFloat(item.metadata.latitude);
    const lng = parseFloat(item.metadata.longitude);
    
    if (isNaN(lat) || isNaN(lng)) {
      return null;
    }
    
    return (
      <Marker
        key={item.id}
        position={{ lat, lng }}
        onClick={() => handleMarkerClick(item)}
        animation={window.google.maps.Animation.DROP}
      />
    );
  };

  if (loadError) {
    return (
      <Alert variant="danger">
        Error loading maps. Please check your internet connection and try again.
      </Alert>
    );
  }

  return isLoaded ? (
    <div className="mb-4">
      <GoogleMap
        mapContainerStyle={containerStyle}
        center={center}
        zoom={12}
        onLoad={onLoad}
        onUnmount={onUnmount}
        options={{
          fullscreenControl: true,
          mapTypeControl: true,
          streetViewControl: true,
          zoomControl: true
        }}
      >
        {items.map(item => renderMarker(item))}
        
        {selectedItem && selectedItem.metadata && (
          <InfoWindow
            position={{
              lat: parseFloat(selectedItem.metadata.latitude),
              lng: parseFloat(selectedItem.metadata.longitude)
            }}
            onCloseClick={() => setSelectedItem(null)}
          >
            <Card style={{ maxWidth: '200px', border: 'none' }}>
              <Card.Body className="p-2">
                <Card.Title className="h6">{selectedItem.name}</Card.Title>
                {selectedItem.description && (
                  <Card.Text className="small text-muted mb-1">
                    {selectedItem.description}
                  </Card.Text>
                )}
                {selectedItem.metadata.address && (
                  <Card.Text className="small">
                    {selectedItem.metadata.address}
                  </Card.Text>
                )}
              </Card.Body>
            </Card>
          </InfoWindow>
        )}
      </GoogleMap>
    </div>
  ) : (
    <div className="text-center my-4">
      <Spinner animation="border" role="status">
        <span className="visually-hidden">Loading map...</span>
      </Spinner>
    </div>
  );
};

export default LocationMapView; 