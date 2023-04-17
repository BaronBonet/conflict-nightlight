import Map, {Source, Layer} from 'react-map-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import {v4 as uuidv4} from 'uuid';

import mapboxgl from "mapbox-gl";
import {useCallback, useEffect, useState} from "react";
import YearMonthSelect from "./YearMonthSelect";
import useWindowDimensions from "../hooks";

// eslint-disable-next-line import/no-webpack-loader-syntax
mapboxgl.workerClass = require("worker-loader!mapbox-gl/dist/mapbox-gl-csp-worker").default;

const nightlightOverview = {
    id: 'nightlight-layer',
    type: 'raster',
    source: 'mapbox',
    paint: {
        'raster-fade-duration': 0,
        "raster-opacity": 0.5
    }
};

const maxMapBounds = [[21.0, 43.3], [41.3, 54.0]]

const getMapWidth = (width) => {
    if (width > 600) {
        return '50vw'
    } else {
        return '100wv'
    }
}

const getMapHeight = (width) => {
    if (width > 600) {
        return '100vh'
    } else {
        return '50vh'
    }
}

export const TheMap = ({selectedMap, mapSide, activeMap, setViewState, setActiveMap, viewState, mapOptions, setMap }) => {
    const onMove = useCallback(evt => setViewState(evt.viewState), [setViewState]);
    const onMoveStart = useCallback(() => setActiveMap(mapSide), [mapSide, setActiveMap]);

    const {width} = useWindowDimensions();
    const [mapWidth, setMapWidth] = useState(getMapWidth(width));
    const [mapHeight, setMapHeight] = useState(getMapHeight(width));
    const [mapKey, setMapKey] = useState(uuidv4())

    useEffect(() => {
        setMapWidth(getMapWidth(width))
        setMapHeight(getMapHeight(width))
        setMapKey(uuidv4())
    }, [width]);

    const mapStyle = {width: mapWidth, height: mapHeight}

    return (
        <div>
            <Map
                key={mapKey}
                {...viewState}
                style={mapStyle}
                mapStyle="mapbox://styles/ericcbonet/cldc06jgu002j01nyu5qeclku"
                mapboxAccessToken={process.env.REACT_APP_MAPBOX_TOKEN}
                maxZoom={8}
                minZoom={4}
                onMove={activeMap === mapSide && onMove}
                onMoveStart={onMoveStart}
                maxBounds={maxMapBounds}
            >
                <YearMonthSelect mapOptions={mapOptions} selectedMap={selectedMap} setMap={setMap}/>
                <Source id="mapbox-terrain" type="raster" url={selectedMap.url} key={selectedMap.key}>
                    <Layer {...nightlightOverview} />
                </Source>
            </Map>
        </div>
    )
}
