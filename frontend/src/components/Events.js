import React, { Component } from 'react'
import PropTypes from 'prop-types'

export default class Events extends Component {
    static propTypes = {
       events: PropTypes.array,
    }
    
    render() {
        return (
            <div>
                <ul>
                    {this.props.events?.map(e => 
                        <li key={e.id} onClick={ev => this.props.onEventSelected(e.id)}>{e.id}</li>)}
                </ul>
            </div>
        )
    }
}
