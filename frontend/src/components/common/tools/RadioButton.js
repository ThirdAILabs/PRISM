import * as React from 'react';
import Radio from '@mui/material/Radio';
import RadioGroup from '@mui/material/RadioGroup';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormControl from '@mui/material/FormControl';
import { styled } from '@mui/material/styles';

const StyledFormControlLabel = styled(FormControlLabel)(({ checked }) => ({
    '& .MuiFormControlLabel-label': {
        fontWeight: 'bold',
        color: checked ? 'rgb(0, 0, 0)' : '#888888',
    },
    marginRight: '2rem !important',
}));

export default function RowRadioButtonsGroup({ selectedSearchType, formControlProps, handleSearchTypeChange }) {
    return (
        <FormControl>
            <RadioGroup
                row
                aria-labelledby="demo-row-radio-buttons-group-label"
                name="row-radio-buttons-group"
                value={selectedSearchType}
                onClick={handleSearchTypeChange}
            >
                {formControlProps.map((formControlProp) => (
                    <StyledFormControlLabel
                        key={formControlProp.value}
                        value={formControlProp.value}
                        control={<Radio />}
                        label={formControlProp.label}
                        checked={selectedSearchType === formControlProp.value}
                    />
                ))}
            </RadioGroup>
        </FormControl>
    );
}