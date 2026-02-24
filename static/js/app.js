document.addEventListener('DOMContentLoaded', () => {
    document.addEventListener('click', (e) => {
        const notifDd = document.getElementById('notifDropdown');
        const userDd = document.getElementById('userDropdown');

        if (notifDd && !e.target.closest('.notif-wrap')) {
            notifDd.classList.remove('open');
        }
        if (userDd && !e.target.closest('.user-menu-wrap')) {
            userDd.classList.remove('open');
        }
    });
});

function toggleNotifications() {
    const dd = document.getElementById('notifDropdown');
    const userDd = document.getElementById('userDropdown');
    if (userDd) userDd.classList.remove('open');

    dd.classList.toggle('open');

    if (dd.classList.contains('open')) {
        loadNotifications();
    }
}

function toggleUserMenu() {
    const dd = document.getElementById('userDropdown');
    const notifDd = document.getElementById('notifDropdown');
    if (notifDd) notifDd.classList.remove('open');

    dd.classList.toggle('open');
}

function getCSRF() {
    const meta = document.querySelector('meta[name="csrf-token"]');
    return meta ? meta.content : '';
}

function loadNotifications() {
    const dd = document.getElementById('notifDropdown');

    fetch('/api/notifications', {
        headers: { 'X-CSRF-Token': getCSRF() }
    })
    .then(r => r.json())
    .then(notifs => {
        if (!notifs || notifs.length === 0) {
            dd.innerHTML = '<div class="notif-empty">Нет уведомлений</div>';
            return;
        }

        dd.innerHTML = notifs.map(n => {
            let text = '';
            let link = '/';
            const p = n.Payload || {};

            if (n.Type === 'new_response') {
                text = `Новый отклик от <strong>${p.user_name || 'Кто-то'}</strong> на «${p.project_title || 'проект'}»`;
                link = '/project/' + (p.project_id || '');
            } else if (n.Type === 'response_accepted') {
                text = `Ваш отклик на «${p.project_title || 'проект'}» принят`;
                link = '/project/' + (p.project_id || '');
            }

            return `<a href="${link}" class="notif-item ${n.Read ? '' : 'unread'}">${text}</a>`;
        }).join('');

        dd.innerHTML += '<button class="notif-mark-read" onclick="markNotificationsRead(event)">Отметить прочитанными</button>';

        fetch('/api/notifications/read', {
            method: 'POST',
            headers: { 'X-CSRF-Token': getCSRF() }
        }).then(() => {
            const badge = document.querySelector('.notif-badge');
            if (badge) badge.remove();
        });
    })
    .catch(() => {
        dd.innerHTML = '<div class="notif-empty">Ошибка загрузки</div>';
    });
}

function markNotificationsRead(e) {
    e.stopPropagation();
    fetch('/api/notifications/read', {
        method: 'POST',
        headers: { 'X-CSRF-Token': getCSRF() }
    }).then(() => {
        const badge = document.querySelector('.notif-badge');
        if (badge) badge.remove();
        document.querySelectorAll('.notif-item.unread').forEach(el => el.classList.remove('unread'));
    });
}

function toggleRole(card) {
    const cb = card.querySelector('input[type="checkbox"]');
    cb.checked = !cb.checked;
    card.classList.toggle('active', cb.checked);
    const countInput = card.querySelector('.role-count-input');
    if (countInput) {
        if (cb.checked && (!countInput.value || countInput.value === '0')) {
            countInput.value = '1';
        }
    }
}

function stepCount(btn, delta) {
    const input = btn.parentElement.querySelector('.role-count-input');
    if (!input) return;
    const val = parseInt(input.value, 10) || 1;
    const next = val + delta;
    if (next >= 1) input.value = next;
}
