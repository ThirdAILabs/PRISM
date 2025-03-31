import { TextField } from "@mui/material";
import PropTypes from 'prop-types';

const TextFieldComponent = (props) => {
    const defaultSx = {
        '& .MuiOutlinedInput-root': {
            borderRadius: '0.5rem',
            '& fieldset': {
                borderWidth: '1px',
            },
            '&:hover fieldset': {
                borderWidth: '1px',
            },
            '&.Mui-focused fieldset': {
                borderWidth: '1px',
            }
        }
    };

    // Merge custom sx props with default sx
    const combinedSx = {
        ...defaultSx,
        ...(props.sx || {})
    };

    return (
        <TextField
            variant="outlined"
            fullWidth
            autoComplete="off"
            {...props}
            sx={combinedSx}
        />
    );
};

TextFieldComponent.propTypes = {
    sx: PropTypes.object
};

export default TextFieldComponent;