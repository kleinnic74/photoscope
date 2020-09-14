import React, { Component } from 'react'
import PropTypes from 'prop-types'
import './TaskList.css'

export default class TaskList extends Component {
    static propTypes = {
        tasks: PropTypes.arrayOf(PropTypes.object)
    }

    render() {
        const tasks = this.props.tasks.map(t => <li key={t.id} className={t.status}>{t.title}</li>)
        return (
            <ul className='TaskList'>
                {tasks}
            </ul>
        )
    }
}
