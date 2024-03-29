// @ts-nocheck
import { browser } from '$app/environment';
import { writable } from 'svelte/store';

// Mapping of group ID to {...group, messages: []} where group is group data from the backend and messages is an array of chat messages from the backend.
//
// This should be updated by various `get` methods (using `fetch` internally) and also by the `WebSocket`.
//
// Most dashboard Svelte components will derive their properties/state from one particular group.
//
// In particular, the top level dashboard page will read the store, and pass the relevant group (as determined by group ID in path) as a normal property to child components.
export const groups = writable({});
// Mapping of user ID to user, where user is user data from the backend.
//
// This should be updated by various `get` methods (using `fetch` internally) and also by the `WebSocket`.
//
// Most dashboard components will derive usernames and statuses directly from this store.
export const users = writable({});

export const userId = writable(null);

// @ts-nocheck
// If `userId` is undefined, gets the currently-logged-in user.
/**
 * @param {undefined|string} [userId]
 */
async function getUser(userId) {
	try {
		let url = `//${location.host}/api/user/`;
		if (userId) {
			url += `${userId}/`;
		}
		const response = await fetch(url);
		const user = await response.json();
		return user;
	} catch (e) {
		return null;
	}
}

export async function refreshGroup(groupId) {
	const group = await getGroup(groupId);
	groups.update((existing) => {
		return {
			...existing,
			[groupId]: { ...group, messages: existing[groupId] ? existing[groupId].messages : [] }
		};
	});
}

/**
 * @param {string} userId
 */
export async function refreshUser(userId) {
	const user = await getUser(userId);
	users.update((existing) => {
		return {
			...existing,
			[userId]: user
		};
	});
}

async function createGroup(name, calendarMode) {
	try {
		const response = await fetch(`//${location.host}/api/group/`, {
			method: 'PATCH',
			body: JSON.stringify({ name, calendarMode })
		});
		const result = await response.json();
		await refreshGroup(result.groupId);
		return result.groupId;
	} catch (e) {
		return null;
	}
}

async function createAvailability(groupId, availability) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupId}/availability/`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(availability)
		});
		if (response.status === 200) {
			console.log('success for creating availability');
		}
	} catch (e) {
		console.error('Error creating availability:', e);
		return null;
	}
}

async function deleteAvailability(groupId, availabilityId) {
	try {
		const response = await fetch(
			`//${location.host}/api/group/${groupId}/availability/${availabilityId}/`,
			{
				method: 'DELETE'
			}
		);
		if (response.ok) {
			console.log('Availability deleted successfully');
		} else {
			console.error('Failed to delete availability');
		}
	} catch (e) {
		console.error('Error deleting availability:', e);
	}
}

async function createTask(groupId, title) {
	try {
		return await fetch(`//${location.host}/api/group/${groupId}/task/`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ title })
		});
	} catch (e) {
		console.error('Error creating task:', e);
		return null;
	}
}

async function deleteTask(groupId, taskId) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupId}/task/${taskId}/`, {
			method: 'DELETE'
		});
		if (response.ok) {
			console.log('Task deleted successfully');
		} else {
			console.error('Failed to delete task');
		}
	} catch (e) {
		console.error('Error deleting task:', e);
	}
}

async function updateTask(groupId, taskId, taskData) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupId}/task/${taskId}/`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(taskData)
		});
		if (response.ok) {
			console.log('Task updated successfully');
			return true;
		} else {
			console.error('Failed to update task, server responded with status:', response.status);
			return false;
		}
	} catch (e) {
		console.error('Error updating task:', e);
		return false;
	}
}

/**
 * @param {number} groupId
 */
async function getGroup(groupId) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupId}/`);
		const group = await response.json();
		return group;
	} catch (e) {
		return null;
	}
}

async function createPoll(groupId, title, options) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupId}/poll/`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ title, options })
		});
		if (response.status === 200) {
			console.log('success for creating poll');
			const group = await getGroup(groupId);
			console.log('poll as: ', group.poll);
		}
	} catch (e) {
		return null;
	}
}

async function updateVotes(groupID, votes) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupID}/poll/`, {
			method: 'PATCH',
			body: JSON.stringify({ votes })
		});
		if (response.status === 200) {
			console.log('success for updating votes');
		}
	} catch (e) {
		return null;
	}
}

async function deletePoll(groupID) {
	try {
		const response = await fetch(`//${location.host}/api/group/${groupID}/poll/`, {
			method: 'DELETE'
		});
		if (response.status === 200) {
			console.log('success for deleting poll');
		}
	} catch (e) {
		return null;
	}
}

async function sendMessage(groupID, content) {
	try {
		return await fetch(`//${location.host}/api/group/${groupID}/chat/`, {
			method: 'PATCH',
			body: JSON.stringify({ content })
		});
	} catch (e) {
		return null;
	}
}

async function updateUserName(userId, newName) {
	try {
		const response = await fetch(`//${location.host}/api/user/`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ userId: userId, name: newName })
		});
		if (response.ok) {
			users.update((allUsers) => {
				if (allUsers[userId]) {
					allUsers[userId].name = newName;
				}
				return allUsers;
			});
			return true;
		} else {
			console.error('Failed to update user name');
			return false;
		}
	} catch (e) {
		console.error('Error updating user name:', e);
		return false;
	}
}

async function updateStatus(status) {
	try {
		const response = await fetch(`/api/user/`, {
			method: 'PATCH',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ status: status })
		});

		if (response.ok) {
			console.log(`Status updated to ${status}`);
			users.update((u) => {
				if (u[userId]) {
					u[userId].status = status;
				}
				return u;
			});
		} else {
			console.error('Failed to update status');
		}
	} catch (error) {
		console.error('Error updating status:', error);
	}
}

/**
 * @param {any[]} messages
 */
function sortMessages(messages) {
	messages.sort((a, b) => a.timestamp - b.timestamp);
}

async function fetchMessages(groupId, start, end) {
	try {
		const response = await fetch(
			`//${location.host}/api/group/${groupId}/chat/?` + new URLSearchParams({ start, end }),
			{
				method: 'GET'
			}
		);
		const result = await response.json();
		if (result.continue == true) {
			const newEnd = result.messages[0].timestamp - 1;
			fetchMessages(groupId, start, newEnd);
		}
		console.log(result);
		groups.update((existing) => {
			if (!(groupId in existing)) {
				existing[groupId] = { messages: [] };
			}
			existing[groupId].messages = existing[groupId].messages.concat(result.messages);
			sortMessages(existing[groupId].messages);
			return existing;
		});
	} catch (e) {
		console.log(e);
		return null;
	}
}

let webSocket;
async function openWebSocket() {
	const webSocketProtocol = location.protocol == 'http:' ? 'ws:' : 'wss:';
	webSocket = new WebSocket(`${webSocketProtocol}//${location.host}/ws/`);
	webSocket.onopen = console.log;
	webSocket.onmessage = (event) => {
		console.log(event);
		const message = JSON.parse(event.data);
		console.log('message', message);
		if (message.group) {
			refreshGroup(message.group.groupId);
		}
		if (message.user) {
			users.update((existing) => {
				return { ...existing, [message.user.userId]: { ...message.user, userId: undefined } };
			});
		}
		if (message.message) {
			groups.update((existing) => {
				if (!(message.message.groupId in existing)) {
					existing[message.message.groupId] = [];
				}
				existing[message.message.groupId].messages.push(message.message);
				sortMessages(existing[message.message.groupId].messages);
				return existing;
			});
		}
	};
	webSocket.onerror = console.log;
	webSocket.onclose = console.log;
}

if (browser) {
	getUser().then((user) => {
		console.log(user);
		userId.set(user.userId);
		users.update((allUsers) => {
			allUsers[user.userId] = user;
			return allUsers;
		});
		openWebSocket();
	});
}

export {
	getUser,
	createGroup,
	createAvailability,
	createTask,
	createPoll,
	updateVotes,
	deletePoll,
	sendMessage,
	fetchMessages,
	getGroup,
	deleteTask,
	deleteAvailability,
	updateTask,
	updateUserName,
	updateStatus
};
