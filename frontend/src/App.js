import React, { useState, useEffect } from "react";
import "./App.css";

import ReactGA from "react-ga4";
import { TheMap } from "./components/TheMap";
import { CircularProgress, Grid } from "@mui/material";
import ColorLegend from "./components/ColorLegend";
import * as d3 from "d3";
import FullScreenToggle from "./components/FullScreenToggle";
import InfoIcon from "@mui/icons-material/Info";
import LocationSelector from "./components/LocationSelect";

console.log(`version ${process.env.REACT_APP_VERSION}`);
ReactGA.initialize(`${process.env.GOOGLE_TAG}`);
ReactGA.send({ hitType: "pageview", page: window.location.pathname });

// Latitudes are horizontal lines that measure distance north or south of the equator
// Longitudes are vertical lines that measure east or west of the meridian in Greenwich, England
let locationOptions = [
  {
    boundsProtoId: 1,
    label: "Ukraine",
    configuration: {
      zoom: { max: 8, min: 4, default: 5.0 },
      bounds: [
        [21.0, 43.3],
        [41.3, 54.5],
      ],
      startingLocation: { latitude: 48.5, longitude: 32.5 },
      mapSelections: {
        leftMapIndex: 7,
        rightMapIndex: -1, // -1 means last index
      },
    },
  },
  {
    boundsProtoId: 2,
    label: "Palistinine Genocide",
    configuration: {
      zoom: { max: 9, min: 5, default: 8.0 },
      bounds: [
        // bottom right corner [longitude, latitude]
        [33.6, 30.05],
        // top left corner [longitude, latitude]
        [37.5, 34.85],
      ],
      startingLocation: { latitude: 32.0, longitude: 35.0 },
      mapSelections: {
        leftMapIndex: 7,
        rightMapIndex: -1, // -1 means last index
      },
    },
  },
];
const IndexOfStartingLocationOption = 1; // Palestine

function App() {
  const startingConfig =
    locationOptions[IndexOfStartingLocationOption].configuration;
  const [viewState, setViewState] = useState({
    latitude: startingConfig.startingLocation.latitude,
    longitude: startingConfig.startingLocation.longitude,
    zoom: startingConfig.zoom.default,
  });

  const [activeMap, setActiveMap] = useState("left");
  const [leftMap, setLeftMap] = useState(null);
  const [rightMap, setRightMap] = useState(null);
  const [mapOptions, setMapOptions] = useState([]);

  const [selectedLocation, setSelectedLocation] = useState(
    locationOptions[IndexOfStartingLocationOption].boundsProtoId,
  );
  const [allMapOptions, setAllMapOptions] = useState([]); // Store all fetched options

  const mapBorder = { zIndex: "0", border: "1px solid rgba(0, 0, 0, 1)" };

  useEffect(() => {
    const fetchMapOptions = async () => {
      try {
        const response = await fetch(
          "https://cdn.conflictnightlight.com/conflict-nightlight-bounded-map-options.json",
        );
        const fetchedMapOptions = await response.json();
        setAllMapOptions(fetchedMapOptions);

        const defaultOptions =
          fetchedMapOptions.find((d) => d.bounds === selectedLocation)
            ?.maps_options || [];
        setMapOptions(defaultOptions);

        if (defaultOptions.length > 0) {
          const config = locationOptions.find(
            (d) => d.boundsProtoId === selectedLocation,
          )?.configuration;
          const leftIndex = config?.mapSelections?.leftMapIndex ?? 7;
          const rightIndex =
            config?.mapSelections?.rightMapIndex === -1
              ? defaultOptions.length - 1
              : config?.mapSelections?.rightMapIndex;

          setLeftMap(defaultOptions[leftIndex]);
          setRightMap(defaultOptions[rightIndex]);
        }
      } catch (error) {
        console.error("Error fetching map options:", error);
      }
    };

    fetchMapOptions();
  }, [selectedLocation]);

  const colorScale = d3
    .scaleLinear()
    .domain([0, 1])
    .interpolate(() => d3.interpolateRgb("#161616", "#929191"));

  if (!leftMap || !rightMap) {
    return (
      <div className="loading-container">
        <CircularProgress color="inherit" />
      </div>
    );
  }

  const handleBoundsChange = (newBounds) => {
    setSelectedLocation(newBounds);
    const newOptions =
      allMapOptions.find((d) => d.bounds === newBounds)?.maps_options || [];
    setMapOptions(newOptions);
    // Get the selected location configuration
    const selectedConfig = locationOptions.find(
      (d) => d.boundsProtoId === newBounds,
    );
    if (selectedConfig) {
      const config = selectedConfig.configuration;

      setViewState({
        latitude: config.startingLocation.latitude,
        longitude: config.startingLocation.longitude,
        zoom: config.zoom.default,
      });
    }

    if (newOptions.length > 0) {
      const config = selectedConfig?.configuration;
      const leftIndex = config?.mapSelections?.leftMapIndex ?? 7;
      const rightIndex =
        config?.mapSelections?.rightMapIndex === -1
          ? newOptions.length - 1
          : config?.mapSelections?.rightMapIndex;

      setLeftMap(newOptions[leftIndex]);
      setRightMap(newOptions[rightIndex]);
    }
  };

  return (
    <div className="App">
      <a
        href="https://ericcbonet.com/posts/conflict-nightlight/"
        target="_blank"
        rel="noopener noreferrer"
      >
        <InfoIcon
          style={{
            position: "absolute",
            top: "20px",
            left: "20px",
            color: "white",
            zIndex: "100",
          }}
        />
      </a>
      <LocationSelector
        selectedLocation={selectedLocation}
        onLocationChange={handleBoundsChange}
        locations={locationOptions}
      />
      <Grid container columnSpacing={0} spacing={0}>
        <Grid item xs={12} sm={6} style={mapBorder}>
          <TheMap
            selectedMap={leftMap}
            mapSide={"left"}
            activeMap={activeMap}
            setViewState={setViewState}
            viewState={viewState}
            setActiveMap={setActiveMap}
            mapOptions={mapOptions}
            setMap={setLeftMap}
            regionSpecificConfigs={
              locationOptions.find((d) => d.boundsProtoId === selectedLocation)
                ?.configuration
            }
          />
        </Grid>
        <Grid item xs={12} sm={6} style={mapBorder}>
          <TheMap
            selectedMap={rightMap}
            mapSide={"right"}
            activeMap={activeMap}
            setViewState={setViewState}
            viewState={viewState}
            setActiveMap={setActiveMap}
            mapOptions={mapOptions}
            setMap={setRightMap}
            regionSpecificConfigs={
              locationOptions.find((d) => d.boundsProtoId === selectedLocation)
                ?.configuration
            }
          />
        </Grid>
      </Grid>
      <FullScreenToggle />
      <ColorLegend colorScale={colorScale} title="Light Intensity" ticks={5} />
    </div>
  );
}

export default App;
