
import React from 'react';
import styles from './UserLogin.module.css'
import monkey from '../../assets/monkey.png'

function UserLogin({ val,changeInput,onclick}) {
    return (
        <div className={styles.popup}>
            <div>Player Name</div>
            <img src={monkey} alt="monkey" />
            <div className={styles.wrapper} >
                <input
                    type="text"
                    placeholder="Enter your username..."
                    required
                    value={val}
                    onChange={changeInput}
                />
                <button onClick={onclick} type="submit">Get Started</button>
            </div>
        </div>

    );
}
export default UserLogin;