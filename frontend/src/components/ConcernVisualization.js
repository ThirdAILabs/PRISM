import * as React from 'react';
import { Speedometer } from './common/tools/Speedometer';
import "../App.css";

const Hover = ({ text }) => {
  return (
    <div className="popover-container">
      <button
        type="button"
        className="btn btn-info btn-circle ml-2"
        style={{
          marginLeft: '5px',
          width: '14px',
          height: '14px',
          padding: '1px 0',
          borderRadius: '7.5px',
          textAlign: 'center',
          fontSize: '8px',
          lineHeight: '1.42857',
          border: '1px solid grey',
          borderWidth: '1px',
          backgroundColor: 'transparent',
          color: 'grey',
          position: 'relative',
          boxShadow: 'none',
        }}
      >
        ?
      </button>
      <div className="popover">
        <div className="popover-body">
          {text}
        </div>
      </div>
    </div>
  );
}


export default function ConcernVisualizer({
  title,
  hoverText,
  value,
  scale,
  weight,
  setWeight,
  onReview,
}) {
  const [weightString, setWeightString] = React.useState(weight.toString());

  const updateWeight = () => {
    let newWeight = 0;
    if (weightString != "") {
      newWeight = parseFloat(weightString);
    }
    setWeight(newWeight);
  }

  return (
    <div className='chart-wrapper' style={{
      position: 'relative'
    }}>
      <Speedometer scale={scale || [0, 1, 2, 3, 5, 10, 20]} value={value} />

      <div className='mt-3 mb-4 text-light' style={{ height: "50px" }} >
        {title} <Hover text={hoverText} />
      </div>

      <div className='mb-3'>
        <input type='field' className='btn btn-dark rounded rounded-5' style={{ height: '50px', width: '50px' }} value={weightString}
          onChange={(e) => {
            if (e.target.value == "") {
              setWeightString("");
            }
            const parsed = parseFloat(e.target.value);
            if (parsed != NaN) {
              setWeightString(e.target.value);
            }
          }}
          onKeyDown={(e) => {
            if (e.key == "Enter") {
              updateWeight();
              e.target.blur();
            }
          }}
          onBlur={updateWeight} />
      </div>

      <button className='btn btn-dark rounded rounded-5 px-4' onClick={onReview}>Review</button>

    </div>
  );
}