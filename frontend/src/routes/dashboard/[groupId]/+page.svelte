<script>
	// @ts-nocheck

	import { onMount } from 'svelte';
	import dayjs from 'dayjs';
	import { get } from 'svelte/store';
	import {
		createAvailability,
		createTask,
		deleteAvailability,
		deleteTask,
		getGroup,
		groups,
		refreshGroup,
		updateTask,
		userId,
		updateUserName,
		users,
		refreshUser,
		updateStatus,
		fetchMessages
	} from '$lib/model';
	import { goto } from '$app/navigation';
	import Chat from './Chat.svelte';
	import { page } from '$app/stores';

	$: groupId = $page.params.groupId;
	$: group = $groups[groupId];
	$: console.log(`group changed: ${JSON.stringify(group)}`);
	$: availability = calculateAvailability($userId, group);
	$: commonAvailability = calculateCommonAvailability(group);

	let isLoadingUsers = true;
	let showMembers = false;
	$: if (group && group.members) {
		group.members.forEach((memberId) => {
			getAssigneeDisplayName(memberId, $users);
		});
	}

	// Bail if the group doesn't exist.
	onMount(async () => {
		await refreshGroup(groupId);
		const group = get(groups)[groupId];
		if (group) {
			fetchMessages(groupId, 0, Number.MAX_SAFE_INTEGER);
			const memberIds = group.members || [];
			await Promise.all(memberIds.map(refreshUser));
			isLoadingUsers = false;
		} else {
			goto('/');
			return;
		}
	});

	function getAssigneeDisplayName(assigneeId, users) {
		const assignee = users[assigneeId];
		return assignee && assignee.name ? assignee.name : `user# ${assigneeId}`;
	}

	function getUserStatus(userId, users) {
		const user = users[userId];
		return user && user.status ? user.status : 'offline';
	}

	let newName = '';
	let isEditingName = false;

	async function handleNameChange() {
		if (newName.trim()) {
			const success = await updateUserName($userId, newName);
			if (success) {
				newName = '';
				isEditingName = false;
			}
		}
	}

	let taskInput = '';
	let isPoll = false;

	function getStartEnd(group) {
		if (!group) {
			return { start: null, end: null };
		}
		const calendarMode = group.calendarMode.split(' to ');
		const dateFormat = 'YYYY-MM-DD';
		const start = dayjs(calendarMode[0], dateFormat);
		const end = dayjs(calendarMode[1], dateFormat);
		return { start, end };
	}

	function calculateAvailability(currentUserId, groupData) {
		if (!currentUserId || !groupData) {
			return {};
		}

		const { start, end } = getStartEnd(groupData);

		if (!(start.isValid() && end.isValid())) {
			return {};
		}

		let availability = {};
		let loopEndDate = end.add(1, 'day');
		for (let current = start; current.isBefore(loopEndDate); current = current.add(1, 'day')) {
			const dateString = current.format('YYYY-MM-DD');
			availability[dateString] = new Array(16).fill(false);
		}

		const userAvailabilities = groupData.availabilities.filter(
			(avail) => avail.userId === currentUserId
		);
		userAvailabilities.forEach(({ date, start }) => {
			const hour = parseInt(start.split(':')[0], 10) - 7;
			if (availability[date]) {
				availability[date][hour] = true;
			}
		});
		return availability;
	}

	function _slotsChanged(newSlots, oldSlots) {
		if (newSlots.length !== oldSlots.length) return true;
		for (let i = 0; i < newSlots.length; i++) {
			if (newSlots[i] !== oldSlots[i]) {
				return true;
			}
		}
		return false;
	}

	function calculateCommonAvailability(groupData) {
		if (!groupData) {
			return { isLoading: true, slots: [] };
		}

		// date -> [{userId, start, end}]
		let availabilityRanges = {};

		groupData.availabilities.forEach(({ userId, date, start, end }) => {
			const startTime = dayjs(`${date} ${start}`);
			const endTime = dayjs(`${date} ${end}`);
			if (!availabilityRanges[date]) {
				availabilityRanges[date] = [];
			}
			availabilityRanges[date].push({
				userId: userId,
				start: startTime,
				end: endTime
			});
		});

		let commonSlots = {};
		Object.keys(availabilityRanges).forEach((date) => {
			let ranges = availabilityRanges[date];
			ranges.sort((a, b) => a.start - b.start);

			let overlapping = [];
			ranges.forEach((currentRange, index) => {
				for (let i = index + 1; i < ranges.length; i++) {
					let nextRange = ranges[i];
					if (currentRange.end > nextRange.start) {
						let startOverlap = nextRange.start.isAfter(currentRange.start)
							? nextRange.start
							: currentRange.start;
						let endOverlap = nextRange.end.isBefore(currentRange.end)
							? nextRange.end
							: currentRange.end;
						if (
							!overlapping.some((ov) => ov.start.isSame(startOverlap) && ov.end.isSame(endOverlap))
						) {
							overlapping.push({ start: startOverlap, end: endOverlap });
						}
					}
				}
			});

			if (overlapping.length) {
				commonSlots[date] = overlapping;
			}
		});

		let formattedCommonSlots = [];
		Object.keys(commonSlots).forEach((date) => {
			commonSlots[date].forEach((slot) => {
				formattedCommonSlots.push(
					`${date} from ${slot.start.format('HH:mm')} to ${slot.end.format('HH:mm')}`
				);
			});
		});

		return { isLoading: false, slots: formattedCommonSlots };
	}

	async function addTask(title) {
		if (!groupId) {
			return;
		}

		try {
			const response = await createTask(groupId, title);
			if (response.ok) {
				taskInput = '';
			} else {
				console.error('server error');
			}
		} catch (e) {
			console.error('task error ', e);
		}
	}

	async function toggleCompletion(taskId, groupData) {
		const task = groupData.tasks.find((t) => t.taskId === taskId);
		if (task) {
			try {
				const newCompletedStatus = !task.completed;
				console.log('status', newCompletedStatus);
				const success = await updateTask(groupId, taskId, { completed: newCompletedStatus });

				if (!success) {
					console.error('Failed to update task completion on server.');
				}
			} catch (error) {
				console.error('Error updating task completion:', error);
			}
		}
	}

	function openPoll() {
		if (isPoll) {
			isPoll = false;
		} else {
			isPoll = true;
		}
	}

	async function addAvailability(date, hour) {
		hour += 7;
		await createAvailability(groupId, {
			date,
			start: `${hour < 10 ? `0${hour}` : hour}:00`,
			end: `${hour + 1 < 10 ? `0${hour + 1}` : hour + 1}:00`
		});
	}

	async function removeAvailability(selectedDay, selectedHour) {
		selectedHour += 7;
		const formattedHour = `${selectedHour < 10 ? `0${selectedHour}` : selectedHour}:00`;
		const currentData = await getGroup(groupId);
		const matchingAvailability = currentData.availabilities.find(
			(avail) => avail.date === selectedDay && avail.start === formattedHour
		);

		console.log(groupId);
		if (matchingAvailability) {
			console.log(
				'making an attempt to delete availability with id: ',
				matchingAvailability.availabilityId
			);
			await deleteAvailability(groupId, matchingAvailability.availabilityId);
			console.log(`Deleted availability with ID: ${matchingAvailability.availabilityId}`);
		} else {
			console.error('No matching availability found to delete');
		}
	}

	async function deleteTaskWrapper(taskId) {
		try {
			await deleteTask(groupId, taskId);
		} catch (error) {
			console.error(error);
		}
	}

	async function assignTaskToUser(taskId) {
		console.log('taskid:', taskId);
		console.log('user id:', userId);
		const taskData = {
			assignee: $userId
		};
		const success = await updateTask(groupId, taskId, taskData);
		if (success) {
			console.log('Task assigned');
		} else {
			console.error('Failed to assign task.');
		}
	}

	$: console.log(availability);
</script>

<header />

<main>
	<div class="content-wrap">
		<div class="menu-bar">
			<button class="menu-button">
				<img src="../menubar.png" alt="menu bar" class="hamburger-icon" />
				<span class="logo">LemmeKnow</span>
			</button>
			<button class="menu-button" on:click={() => (showMembers = !showMembers)}>
				<img src="../users.png" alt="menu bar" class="user-icon" />
				<span class="members-title">Members</span>
			</button>
			<button class="menu-button" on:click={openPoll}>
				<img src="../vote.png" alt="menu bar" class="poll-icon" />
				<span class="members-title-poll">Poll</span>
			</button>
			<button
				on:click={() => {
					navigator.clipboard.writeText(`${window.location.origin}/dashboard/${groupId}`);
					document.querySelector('.invite-button').innerText = 'Copied to Clipboard!';
					setTimeout(() => {
						document.querySelector('.invite-button').innerText = 'Copy Invite Link!';
					}, 1500);
				}}
				class="invite-button"
				>Copy Invite Link!
			</button>
			{#if isEditingName}
				<form on:submit|preventDefault={handleNameChange}>
					<input
						class="name-input"
						type="text"
						bind:value={newName}
						placeholder="Enter your name"
					/>
					<button type="submit" class="update-button">Update Name</button>
					<button type="button" on:click={() => (isEditingName = false)}>Cancel</button>
				</form>
			{:else}
				<button on:click={() => (isEditingName = true)} class="name-button">Change Name</button>
				{#if showMembers}
					<div class="members-list">
						{#if group && group.members}
							<ul>
								<li>Members:</li>
								{#each group.members as memberId (memberId)}
									<li>
										{getAssigneeDisplayName(memberId, $users)}
										<span class="status-button">{getUserStatus(memberId, $users)}</span>
									</li>
									<li>
										{#if memberId === $userId}
											<select on:change={(event) => updateStatus(event.target.value)}>
												<option value="online">Online</option>
												<option value="busy">Busy</option>
												<option value="offline">Offline</option>
											</select>
										{/if}
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				{/if}
			{/if}
		</div>

		<Chat {groupId} {group} bind:isPoll />

		<div class="calendar-container">
			{#if $users[$userId] && $users[$userId].name}
				<div class="name-display">
					<h2>Hello, {$users[$userId].name}!</h2>
				</div>
			{/if}
			{#if group}
				<div class="group-name-display">
					<h2>Welcome to "{group.name}"!</h2>
				</div>
			{/if}
			<span class="calendar-title">AVAILABILITY CALENDAR</span>
			{#each Object.keys(availability) as day}
				<div class="day">
					<h3>{day}</h3>
					<div class="slots">
						{#each availability[day] as available, hour}
							<div
								class="slot {available ? 'available' : ''}"
								on:click|preventDefault={() => !available && addAvailability(day, hour)}
								on:keypress={() => !available && addAvailability(day, hour)}
							>
								{hour + 7}:00
								{#if available}
									<button on:click|preventDefault={() => removeAvailability(day, hour)}
										>Delete</button
									>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			{/each}
			{#if commonAvailability.isLoading}
				<div class="common-availability-message">CALCULATING COMMON AVAILABILITIES...</div>
			{:else if commonAvailability.slots.length > 0}
				<div class="common-availability-message">
					<strong>Common Availabilities:</strong>
					<ul>
						{#each commonAvailability.slots as slot}
							<li>{slot}</li>
						{/each}
					</ul>
				</div>
			{:else}
				<div class="common-availability-message">NO COMMON AVAILABILITIES FOUND.</div>
			{/if}
			<form on:submit|preventDefault={() => addTask(taskInput)}>
				<input
					type="text"
					bind:value={taskInput}
					placeholder="Enter task description (50 characters max)"
					maxlength="50"
					style="margin-bottom: -3rem;"
				/>
				<button type="submit" disabled={!taskInput.trim()}>Add Task</button>
			</form>
			{#each group ? group.tasks : [] as task (task.taskId)}
				{#if isLoadingUsers}
					<p>Loading...</p>
				{:else}
					<div class="task-item">
						<input
							type="checkbox"
							bind:checked={task.completed}
							on:click={() => toggleCompletion(task.taskId, group)}
							on:keypress={() => toggleCompletion(task.taskId, group)}
						/>
						<span class={task.completed ? 'completed-task' : ''}>{task.title}</span>
						{#if task.assignee}
							<span class={task.completed ? 'completed-task' : 'incomplete-task'}
								>Assigned to: {getAssigneeDisplayName(task.assignee, $users)}</span
							>
						{/if}
						<button class="delete-task" on:click={() => deleteTaskWrapper(task.taskId)}
							>delete</button
						>
						<button class="self-assign" on:click={() => assignTaskToUser(task.taskId)}
							>Self Assign</button
						>
					</div>
				{/if}
			{/each}
		</div>
	</div>
</main>

<style>
	.calendar-title {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		flex-direction: column;
		text-align: center;
		font-size: 3rem;
		font-family: 'Baloo Bhai 2';
		margin-left: 0.5rem;
		font-weight: bolder;
		color: black;
	}
	.name-display {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		flex-direction: column;
		text-align: center;
		font-size: 1.5rem;
		font-family: 'Baloo Bhai 2';
		margin-left: 2rem;
		margin-top: -1rem;
		margin-bottom: -2.5rem; /* Adjusted margin-bottom */
		font-weight: bolder;
		color: black;
	}

	.group-name-display {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		flex-direction: column;
		text-align: center;
		font-size: 1.25rem;
		font-family: 'Baloo Bhai 2';
		margin-left: 2rem;
		margin-top: -1rem;
		margin-bottom: -2rem; /* Adjusted margin-bottom */
		font-weight: bolder;
		color: black;
	}

	.menu-bar {
		position: relative;
		display: flex;
		flex-direction: column;
		top: 0;
		left: 0;
		align-items: flex-start;
	}

	.menu-bar form {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		margin-left: 0.8rem;
		margin-bottom: 1rem;
		width: 90%;
	}

	.user-icon {
		width: 3rem;
		display: block;
		margin-left: 1.5rem;
	}
	.poll-icon {
		width: 3rem;
		display: block;
		margin-left: 2.5rem;
	}

	.menu-button {
		background: none;
		border: none;
		cursor: pointer;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		line-height: 1;
		text-align: center;
	}

	.menu-button:focus {
		outline: none;
	}

	.logo {
		font-size: 1.5rem;
		margin-top: 0.25rem;
		font-family: 'Baloo Bhai 2';
		font-weight: bolder;
		color: #73a0e7;
	}

	img.hamburger-icon {
		width: 3rem;
		margin: 0 auto;
		display: block;
	}

	.members-title {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		text-align: center;
		font-size: 1.5rem;
		margin-top: 0.25rem;
		margin-left: 1rem;
		font-family: 'Baloo Bhai 2';
		font-weight: bolder;
		color: black;
	}
	.members-title-poll {
		display: flex;
		align-items: flex-start;
		justify-content: center;
		text-align: center;
		font-size: 1.5rem;
		margin-top: 0.25rem;
		margin-left: 2rem;
		font-family: 'Baloo Bhai 2';
		font-weight: bolder;
		color: black;
	}

	.members-list ul {
		list-style: none;
		padding: 0;
	}

	.members-list li {
		margin-left: 1rem;
		display: inline-list-item;
		font-family: 'Baloo Da 2';
		font-weight: bold;
		font-size: large;
	}

	.menu-button:hover .hamburger-icon {
		transform: scale(1.2);
	}

	.menu-button:hover .user-icon {
		transform: scale(1.2);
	}

	.menu-button:hover .poll-icon {
		transform: scale(1.2);
	}

	.content-wrap {
		display: flex;
		flex-direction: row;
		gap: 2rem;
	}

	.calendar-container {
		display: flex;
		flex-direction: column;
		flex-wrap: wrap;
		margin-left: 2rem;
		margin-bottom: 0.15rem;
	}

	.day {
		border: 1px solid #ccc;
		padding: 10px;
	}

	.slots {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 5px;
	}

	.slot {
		padding: 5px;
		background-color: #f0f0f0;
		text-align: center;
		cursor: pointer;
	}

	.slot.available {
		background-color: #a9e1a9;
	}

	input {
		padding: 0.5rem;
		margin-bottom: 0.5rem;
		margin-top: 0.5rem;
		width: 80%;
		max-width: 25rem;
		text-align: center;
		font-size: 1rem;
		background-color: #c9e7e7;
		border-radius: 10px;
		border: 2px solid transparent;
	}

	:global(button) {
		padding: 0.5rem 1rem;
		background-color: #2774d0;
		color: white;
		border: none;
		align-items: center;
		font-size: 1rem;
		border-radius: 1rem;
		cursor: pointer;
	}

	:global(button:hover) {
		background-color: gray;
		color: white;
	}

	.task-item {
		display: flex;
		padding: 0.5rem;
		background-color: #f9f9f9;
		border-radius: 0.2rem;
		font-family: 'Baloo Bhai 2';
		align-items: center;
		font-size: 1rem;
		margin-top: 0;
		margin-bottom: 0.5rem;
	}

	.task-item input[type='checkbox'] {
		accent-color: #879db7;
		transform: scale(1.5);
		cursor: pointer;
		margin-top: -1rem;
		margin-left: -4rem;
		margin-right: -4rem;
	}

	.task-item .completed-task {
		text-decoration: line-through;
		color: #879db7;
	}

	.task-item span {
		margin-right: 1rem;
		color: #333;
		text-align: center;
		font-weight: bold;
	}

	:global(button[type='submit']:disabled) {
		background-color: #ccc;
		cursor: not-allowed;
	}

	:global(input) {
		padding: 0.5rem;
		margin-bottom: 0.5rem;
		width: 80%;
		max-width: 300px;
		text-align: center;
		font-size: 1rem;
		background-color: #eedaf4;
		border-radius: 15px;
		border: 2px solid transparent;
	}

	:global(button) {
		flex-shrink: 0;
	}

	.invite-button {
		display: block;
		margin: 1rem auto;
		background-color: #76a6e7;
		font-weight: bolder;
		font-family: 'Baloo Bhai 2';
		font-size: large;
		color: black;
	}

	.invite-button:hover {
		background-color: #afaeae;
		color: white;
	}

	.delete-task {
		background-color: #879db7;
		color: black;
		border: none;
		cursor: pointer;
		margin-left: 1.5rem;
		padding: 0.5rem 1rem;
		display: inline-block;
		text-align: center;
		font-size: 1rem;
		border-radius: 0.3rem;
	}

	.delete-task:hover {
		background-color: gray;
		color: white;
	}

	.self-assign {
		background-color: #879db7;
		color: black;
		border: none;
		cursor: pointer;
		margin-left: 1.5rem;
		padding: 0.5rem 1rem;
		display: inline-block;
		text-align: center;
		font-size: 1rem;
		border-radius: 0.3rem;
	}

	.self-assign:hover {
		background-color: gray;
		color: white;
	}

	.common-availability-message {
		text-align: left;
		font-family: 'Baloo Da 2';
		font-style: oblique;
		margin-top: 1rem;
		font-size: 1.25rem;
	}

	.save-avail {
		background-color: #87c7fa;
		color: black;
		border: none;
		font-weight: bold;
		cursor: pointer;
		margin-top: 0.25rem;
		padding: 0.5rem;
		display: inline-block;
		text-align: center;
		font-size: 1rem;
		border-radius: 0.3rem;
		margin-bottom: 1rem;
	}

	.save-avail:hover {
		background-color: gray;
		color: white;
	}

	.name-button {
		display: block;
		margin: 0.3rem auto;
		background-color: #76a6e7;
		font-weight: bolder;
		font-family: 'Baloo Bhai 2';
		font-size: large;
		color: black;
	}

	.name-button:hover {
		background-color: #afaeae;
		color: white;
	}

	.update-button {
		margin-bottom: 0.25rem;
	}

	.status-button {
		color: red;
	}
</style>
