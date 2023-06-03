import React, { useState, useEffect } from 'react';
import './App.css';

import ReactGA from "react-ga4";
import { TheMap } from './components/TheMap';
import {CircularProgress, Grid} from '@mui/material';
import ColorLegend from './components/ColorLegend'
import * as d3 from 'd3';
import FullScreenToggle from "./components/FullScreenToggle";
import InfoIcon from '@mui/icons-material/Info';

console.log(`version ${process.env.REACT_APP_VERSION}`);

function App() {
    const [viewState, setViewState] = useState({
        latitude: 48.5,
        longitude: 31.4,
        zoom: 5.4,
    });

    const [activeMap, setActiveMap] = useState('left');
    const [leftMap, setLeftMap] = useState(null);
    const [rightMap, setRightMap] = useState(null);
    const [mapOptions, setMapOptions] = useState([]);

    const mapBorder = {'zIndex': "0", border: "1px solid rgba(0, 0, 0, 1)"}

    ReactGA.initialize(`${process.env.GOOGLE_TAG}`);

    useEffect(() => {
        const fetchMapOptions = async () => {
            try {
                const response = await fetch('https://cdn.conflictnightlight.com/conflict-nightlight-map-options.json');
                const fetchedMapOptions = await response.json();

                setMapOptions(fetchedMapOptions);
                // TODO set this through the backend
                setLeftMap(fetchedMapOptions[7]);
                setRightMap(fetchedMapOptions[fetchedMapOptions.length -1]);
            } catch (error) {
                console.error('Error fetching map options:', error);
            }
        };

        fetchMapOptions();
    }, []);

    const colorScale = d3.scaleLinear().domain([0, 1]).interpolate(() => d3.interpolateRgb('#161616', '#929191'));


    if (!leftMap || !rightMap) {
        return (
            <div className="loading-container">
                <CircularProgress color="inherit" />
            </div>
        );
    }

    return (
        <div className="App">
            <a href="https://ericcbonet.com/posts/conflict-nightlight/" target="_blank" rel="noopener noreferrer">
                <InfoIcon style={{ position: 'absolute', top: '20px', left: '20px', color: 'white', zIndex: "100" }} />
            </a>
            <Grid container columnSpacing={0} spacing={0}>
                <Grid item xs={12} sm={6} style={mapBorder}>
                    <TheMap
                        selectedMap={leftMap}
                        mapSide={'left'}
                        activeMap={activeMap}
                        setViewState={setViewState}
                        viewState={viewState}
                        setActiveMap={setActiveMap}
                        mapOptions={mapOptions}
                        setMap={setLeftMap}
                    />
                </Grid>
                <Grid item xs={12} sm={6} style={mapBorder}>
                    <TheMap
                        selectedMap={rightMap}
                        mapSide={'right'}
                        activeMap={activeMap}
                        setViewState={setViewState}
                        viewState={viewState}
                        setActiveMap={setActiveMap}
                        mapOptions={mapOptions}
                        setMap={setRightMap}
                    />
                </Grid>
            </Grid>
            <FullScreenToggle />
            <ColorLegend colorScale={colorScale} title="Light Intensity" ticks={5}/>
        </div>
    );
}

export default App;
