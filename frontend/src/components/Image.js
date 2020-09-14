import React from 'react'
import PropTypes from 'prop-types'

function Image(props) {
    return (
        <img src={props.src} alt={props.alt} onClick={props.onClick} />
    )
}

Image.propTypes = {
    src: PropTypes.string,
    alt: PropTypes.string,
    onClick: PropTypes.func
}

export default Image

