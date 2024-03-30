// Import the functions you need from the SDKs you need
import { initializeApp } from 'https://www.gstatic.com/firebasejs/10.5.2/firebase-app.js';
import { getAuth, signInWithPopup, GoogleAuthProvider } from 'https://www.gstatic.com/firebasejs/10.5.2/firebase-auth.js';

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
	apiKey: 'AIzaSyDgdYkUpixss87VF6DbAKle4LKQG_eXs_k',
	authDomain: 'bigbox-34654.firebaseapp.com',
	projectId: 'bigbox-34654',
	storageBucket: 'bigbox-34654.appspot.com',
	messagingSenderId: '441356283577',
	appId: '1:441356283577:web:cce7891fec089402bbb568',
	measurementId: 'G-361PPX2KVD'
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
const auth = getAuth(app);

const provider = new GoogleAuthProvider();
window.signInOrOut = () => {
	if (auth.currentUser.isAnonymous) {
		signInWithPopup(auth, provider);
	} else {
		auth.signOut();
	}
};

window.auth = auth;
