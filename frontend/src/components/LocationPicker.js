import React, { useState, useCallback, useEffect } from 'react';
import { GoogleMap, useJsApiLoader, Marker, StandaloneSearchBox } from '@react-google-maps/api';
import { Form, InputGroup, Button, Spinner, Alert } from 'react-bootstrap';
import { FaSearch, FaMapMarkerAlt } from 'react-icons/fa';

const containerStyle = {
  width: '100%',
  height: '300px',
  borderRadius: '8px'
};

const defaultCenter = {
  lat: 37.7749, // San Francisco default
  lng: -122.4194
};

// Google Maps API key should be stored in environment variable
const apiKey = process.env.REACT_APP_GOOGLE_MAPS_API_KEY || '';

const libraries = ['places'];

const LocationPicker = ({ 
  value = { 
    latitude: '', 
    longitude: '', 
    address: '',
    location: ''
  }, 
  onChange 
}) => {
  const [map, setMap] = useState(null);
  const [searchBox, setSearchBox] = useState(null);
  const [center, setCenter] = useState(defaultCenter);
  const [markerPosition, setMarkerPosition] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [isGeocoding, setIsGeocoding] = useState(false);

  const { isLoaded, loadError } = useJsApiLoader({
    id: 'google-map-script',
    googleMapsApiKey: apiKey,
    libraries
  });

  // Initialize marker from value if provided
  useEffect(() => {
    if (isLoaded && value) {
      if (value.latitude && value.longitude) {
        const lat = parseFloat(value.latitude);
        const lng = parseFloat(value.longitude);
        
        if (!isNaN(lat) && !isNaN(lng)) {
          setMarkerPosition({ lat, lng });
          setCenter({ lat, lng });
        }
      }
      
      if (value.location) {
        setSearchTerm(value.location);
      }
    }
  }, [isLoaded, value]);

  const onMapLoad = useCallback(map => {
    setMap(map);
  }, []);

  const onUnmount = useCallback(() => {
    setMap(null);
  }, []);

  const onSearchBoxLoad = useCallback(searchBox => {
    setSearchBox(searchBox);
  }, []);

  const onPlacesChanged = useCallback(() => {
    if (searchBox) {
      const places = searchBox.getPlaces();
      
      if (places.length === 0) return;
      
      const place = places[0];
      
      if (!place.geometry || !place.geometry.location) return;
      
      // Update marker and center
      const newPosition = {
        lat: place.geometry.location.lat(),
        lng: place.geometry.location.lng()
      };
      
      setMarkerPosition(newPosition);
      setCenter(newPosition);
      
      // Get formatted address
      const address = place.formatted_address || '';
      const name = place.name || '';
      
      // Update the form values
      if (onChange) {
        onChange({
          latitude: newPosition.lat.toString(),
          longitude: newPosition.lng.toString(),
          address: address,
          location: name
        });
      }
      
      // Update search term
      setSearchTerm(name);
    }
  }, [searchBox, onChange]);

  const handleMapClick = useCallback(async (e) => {
    const newPosition = {
      lat: e.latLng.lat(),
      lng: e.latLng.lng()
    };
    
    setMarkerPosition(newPosition);
    
    // Reverse geocode to get address
    if (window.google && window.google.maps) {
      setIsGeocoding(true);
      
      try {
        const geocoder = new window.google.maps.Geocoder();
        const result = await geocoder.geocode({ location: newPosition });
        
        if (result.results && result.results.length > 0) {
          const place = result.results[0];
          const address = place.formatted_address || '';
          
          // Find a name for the location (often a business name or landmark)
          let name = '';
          if (place.name) {
            name = place.name;
          } else {
            // Try to get a meaningful name from address components
            const nameComponents = place.address_components.filter(
              comp => comp.types.includes('point_of_interest') || 
                      comp.types.includes('establishment') ||
                      comp.types.includes('premise')
            );
            
            if (nameComponents.length > 0) {
              name = nameComponents[0].long_name;
            } else {
              name = address.split(',')[0] || 'Selected Location';
            }
          }
          
          setSearchTerm(name);
          
          // Update the form values
          if (onChange) {
            onChange({
              latitude: newPosition.lat.toString(),
              longitude: newPosition.lng.toString(),
              address: address,
              location: name
            });
          }
        }
      } catch (err) {
        console.error('Error geocoding:', err);
      } finally {
        setIsGeocoding(false);
      }
    }
  }, [onChange]);

  const handleSearchInputChange = (e) => {
    setSearchTerm(e.target.value);
  };

  const handleCurrentLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        async (position) => {
          const newPosition = {
            lat: position.coords.latitude,
            lng: position.coords.longitude
          };
          
          setMarkerPosition(newPosition);
          setCenter(newPosition);
          
          // Reverse geocode to get address (reusing code from handleMapClick)
          if (window.google && window.google.maps) {
            setIsGeocoding(true);
            
            try {
              const geocoder = new window.google.maps.Geocoder();
              const result = await geocoder.geocode({ location: newPosition });
              
              if (result.results && result.results.length > 0) {
                const place = result.results[0];
                const address = place.formatted_address || '';
                const name = place.name || address.split(',')[0] || 'Current Location';
                
                setSearchTerm(name);
                
                // Update the form values
                if (onChange) {
                  onChange({
                    latitude: newPosition.lat.toString(),
                    longitude: newPosition.lng.toString(),
                    address: address,
                    location: name
                  });
                }
              }
            } catch (err) {
              console.error('Error geocoding:', err);
            } finally {
              setIsGeocoding(false);
            }
          }
        },
        (error) => {
          console.error('Error getting location:', error);
          alert('Unable to get your location. Please make sure location services are enabled.');
        }
      );
    } else {
      alert('Geolocation is not supported by your browser.');
    }
  };

  if (loadError) {
    return (
      <Alert variant="danger">
        Error loading maps. Please check your internet connection and try again.
      </Alert>
    );
  }

  return (
    <div className="location-picker">
      <Form.Group className="mb-3">
        <Form.Label>Location Name</Form.Label>
        <InputGroup>
          {isLoaded && (
            <StandaloneSearchBox
              onLoad={onSearchBoxLoad}
              onPlacesChanged={onPlacesChanged}
            >
              <Form.Control
                type="text"
                placeholder="Search for a place or address"
                value={searchTerm}
                onChange={handleSearchInputChange}
              />
            </StandaloneSearchBox>
          )}
          <Button 
            variant="outline-secondary" 
            onClick={handleCurrentLocation}
            title="Use current location"
          >
            <FaMapMarkerAlt />
          </Button>
        </InputGroup>
        <Form.Text className="text-muted">
          Search for a place or click on the map to set a location
        </Form.Text>
      </Form.Group>

      {isLoaded ? (
        <GoogleMap
          mapContainerStyle={containerStyle}
          center={center}
          zoom={14}
          onLoad={onMapLoad}
          onUnmount={onUnmount}
          onClick={handleMapClick}
          options={{
            fullscreenControl: false,
            mapTypeControl: true,
            streetViewControl: false,
            zoomControl: true
          }}
        >
          {markerPosition && (
            <Marker
              position={markerPosition}
              draggable={true}
              onDragEnd={handleMapClick}
            />
          )}
        </GoogleMap>
      ) : (
        <div style={containerStyle} className="d-flex justify-content-center align-items-center bg-light">
          <Spinner animation="border" role="status">
            <span className="visually-hidden">Loading map...</span>
          </Spinner>
        </div>
      )}

      {/* Hidden fields for form submission */}
      <input type="hidden" value={value.latitude || ''} name="latitude" />
      <input type="hidden" value={value.longitude || ''} name="longitude" />
      <input type="hidden" value={value.address || ''} name="address" />

      {isGeocoding && (
        <div className="text-center my-2">
          <Spinner animation="border" size="sm" role="status">
            <span className="visually-hidden">Finding address...</span>
          </Spinner>
          <span className="ms-2 text-muted">Finding address information...</span>
        </div>
      )}
    </div>
  );
};

export default LocationPicker; 