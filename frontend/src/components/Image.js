import React from 'react'
import PropTypes from 'prop-types'

function Image(props) {
    console.log("Image", props)
    return (
        <img src={props.src} alt={props.alt}/>
    )
}

Image.propTypes = {
    src: PropTypes.string,
    alt: PropTypes.string
}

export default Image

