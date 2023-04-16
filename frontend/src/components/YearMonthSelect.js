import * as React from 'react';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';

export default function YearMonthSelect({mapOptions, selectedMap, setMap}) {

    const handleChange = (event) => {
        setMap(event.target.value);
    };


    return (
        <div className='year-month-select-div'>
                    <FormControl className='year-month-select'  sx={{m: 1, minWidth: 120}}>
                        <Select
                            value={selectedMap}
                            onChange={handleChange}
                            autoWidth
                        >
                            {mapOptions.map((object, i) => <MenuItem value={object} key={i}>{object.display_name}</MenuItem>)}
                        </Select>
                    </FormControl>
        </div>
    );
}
