import React from 'react'
import PropTypes from 'prop-types'

function Navbar(props) {
    var links = []
    for (const [name, l] of Object.entries(props.links)) {
        links.push(<button key={l.href} onClick={(ev) => props.onClick(l.href)}>{name}</button>)
    }
    return (
        <div className="navbav">
            {links}
        </div>
    )
}

Navbar.propTypes = {
    links: PropTypes.object
}
Navbar.defaultProps = {
    links: {}
}
export default Navbar
