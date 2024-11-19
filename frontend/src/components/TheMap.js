import Map, { Source, Layer } from "react-map-gl";
import "mapbox-gl/dist/mapbox-gl.css";

import mapboxgl from "mapbox-gl";
import { useCallback, useEffect, useState } from "react";
import YearMonthSelect from "./YearMonthSelect";
import useWindowDimensions from "../hooks";

// prettier-ignore
/* eslint-disable-next-line import/no-webpack-loader-syntax */
mapboxgl.workerClass = require("worker-loader!mapbox-gl/dist/mapbox-gl-csp-worker").default;

const nightlightOverview = {
  id: "nightlight-layer",
  type: "raster",
  source: "mapbox",
  paint: {
    "raster-fade-duration": 0,
    "raster-opacity": 0.5,
  },
};

const getMapWidth = (width) => {
  if (width > 600) {
    return "50vw";
  } else {
    return "100wv";
  }
};

const getMapHeight = (width) => {
  if (width > 600) {
    return "100vh";
  } else {
    return "50vh";
  }
};

export const TheMap = ({
  selectedMap,
  mapSide,
  activeMap,
  setViewState,
  setActiveMap,
  viewState,
  mapOptions,
  setMap,
  regionSpecificConfigs,
}) => {
  const onMove = useCallback(
    (evt) => setViewState(evt.viewState),
    [setViewState],
  );
  const onMoveStart = useCallback(
    () => setActiveMap(mapSide),
    [mapSide, setActiveMap],
  );

  const { width } = useWindowDimensions();
  const [mapWidth, setMapWidth] = useState(getMapWidth(width));
  const [mapHeight, setMapHeight] = useState(getMapHeight(width));

  useEffect(() => {
    setMapWidth(getMapWidth(width));
    setMapHeight(getMapHeight(width));
  }, [width]);

  const mapStyle = { width: mapWidth, height: mapHeight };

  return (
    <div>
      <Map
        key={selectedMap.key}
        {...viewState}
        style={mapStyle}
        mapStyle="mapbox://styles/ericcbonet/cldc06jgu002j01nyu5qeclku"
        mapboxAccessToken={process.env.REACT_APP_MAPBOX_TOKEN}
        maxZoom={regionSpecificConfigs.zoom.max}
        minZoom={regionSpecificConfigs.zoom.min}
        onMove={activeMap === mapSide && onMove}
        onMoveStart={onMoveStart}
        maxBounds={regionSpecificConfigs.bounds}
      >
        <YearMonthSelect
          mapOptions={mapOptions}
          selectedMap={selectedMap}
          setMap={setMap}
        />
        <Source
          id="mapbox-terrain"
          type="raster"
          url={selectedMap.url}
          key={selectedMap.key}
        >
          <Layer {...nightlightOverview} />
        </Source>
      </Map>
    </div>
  );
};
