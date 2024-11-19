import React from "react";
import { FormControl, Select, MenuItem } from "@mui/material";

const LocationSelector = ({
  selectedLocation,
  onLocationChange,
  locations,
}) => {
  return (
    <div className="location-select-div">
      <FormControl
        className="selector"
        sx={{ m: 1, minWidth: 120 }}
        size="small"
      >
        <Select
          value={selectedLocation}
          onChange={(e) => onLocationChange(e.target.value)}
          label="Select Region"
        >
          {locations.map((location) => (
            <MenuItem
              key={location.boundsProtoId}
              value={location.boundsProtoId}
            >
              {location.label}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </div>
  );
};

export default LocationSelector;
