import React from 'react'
import PropTypes from 'prop-types'

function Navbar(props) {
    if (props.links) {
        const links = props.links.map(l => (
            <button onClick={(ev) => props.onClick(l.href)}>{l.name}</button>
        ))
        return (
            <div className="navbav">
                {links}
            </div>
        )
    } else {
        return (<div className="navbar"></div>)
    }
}

Navbar.propTypes = {
    links: PropTypes.arrayOf(PropTypes.object)
}
Navbar.defaultProps = {
    links: []
}
export default Navbar
