import React, {useState} from 'react';
import FullscreenIcon from '@mui/icons-material/Fullscreen';
import FullscreenExitIcon from '@mui/icons-material/FullscreenExit';
import IconButton from '@mui/material/IconButton';

const FullscreenToggle = () => {
    const [isFullscreen, setIsFullscreen] = useState(false);

    const toggleFullscreen = () => {
        if (!document.fullscreenElement) {
            document.documentElement.requestFullscreen();
            setIsFullscreen(true);
        } else {
            if (document.exitFullscreen) {
                document.exitFullscreen();
                setIsFullscreen(false);
            }
        }
    };

    return (
        <IconButton onClick={toggleFullscreen} color="inherit" style={{ color:'white', position: 'absolute', bottom: '20px', left: '20px', zIndex: 10}}>
            {isFullscreen ? <FullscreenExitIcon /> : <FullscreenIcon />}
        </IconButton>
    );
};

export default FullscreenToggle;
