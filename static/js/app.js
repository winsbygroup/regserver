// Configure HTMX to include CSRF token in all requests
document.addEventListener('DOMContentLoaded', function() {
	var csrfMeta = document.querySelector('meta[name="csrf-token"]');
	if (csrfMeta) {
		var csrfToken = csrfMeta.getAttribute('content');
		if (csrfToken) {
			document.body.addEventListener('htmx:configRequest', function(evt) {
				evt.detail.headers['X-CSRF-Token'] = csrfToken;
			});
		}
	}
});

// Theme toggle: Light <-> Dark
function toggleTheme() {
	var html = document.documentElement;
	var current = html.getAttribute('data-theme');
	var next = current === 'dark' ? 'light' : 'dark';

	// Update theme
	html.setAttribute('data-theme', next);

	// Save to cookie (1 year expiry)
	document.cookie = 'theme=' + next + ';path=/;max-age=31536000;SameSite=Lax';

	// Update UI
	updateThemeUI(next);
}

function updateThemeUI(theme) {
	var iconLight = document.getElementById('theme-icon-light');
	var iconDark = document.getElementById('theme-icon-dark');
	var label = document.getElementById('theme-label');
	if (!iconLight || !iconDark || !label) return;

	if (theme === 'dark') {
		// In dark mode: show sun icon, offer "Light"
		iconLight.classList.remove('hidden');
		iconDark.classList.add('hidden');
		label.textContent = 'Light';
	} else {
		// In light mode: show moon icon, offer "Dark"
		iconLight.classList.add('hidden');
		iconDark.classList.remove('hidden');
		label.textContent = 'Dark';
	}
}

// Initialize theme UI on page load
document.addEventListener('DOMContentLoaded', function() {
	var theme = document.documentElement.getAttribute('data-theme') || 'light';
	updateThemeUI(theme);

	// Initialize Tom Select on customer dropdown
	var customerSelect = document.getElementById('customer-select');
	if (customerSelect && typeof TomSelect !== 'undefined') {
		var preselected = customerSelect.getAttribute('data-selected');
		var ts = new TomSelect(customerSelect, {
			create: false,
			sortField: { field: 'text', direction: 'asc' },
			onChange: function(value) {
				loadCustomerRegistrations(value);
			}
		});

		// If customer was pre-selected via URL param, load their registrations
		if (preselected) {
			ts.setValue(preselected, true); // true = silent (don't trigger onChange yet)
			loadCustomerRegistrations(preselected);
		}
	}
});

// Helper to show modal
function showModal() {
	document.getElementById('modal').showModal();
}

// Helper to close modal
function closeModal() {
	document.getElementById('modal').close();
}

// Handle modal close (X button or backdrop click)
// If modal content has a data-back-url, navigate back instead of closing
function handleModalClose() {
	var modalContent = document.getElementById('modal-content');
	var backUrl = modalContent.getAttribute('data-back-url');
	if (backUrl) {
		htmx.ajax('GET', backUrl, {target: '#modal-content', swap: 'innerHTML'});
	} else {
		closeModal();
	}
}

// Helper to show toast notification
function showToast(message, type) {
	type = type || 'info';
	const alertClasses = {
		'success': 'alert-success',
		'error': 'alert-error',
		'warning': 'alert-warning',
		'info': 'alert-info'
	};
	const durations = {
		'success': 3000,
		'error': 5000,
		'warning': 5000,
		'info': 3000
	};
	const alertClass = alertClasses[type] || 'alert-info';
	const duration = durations[type] || 3000;

	// Check if modal is open - if so, show toast inside modal to escape dialog top layer
	const modal = document.getElementById('modal');
	const isModalOpen = modal && modal.open;

	const toast = document.createElement('div');
	toast.className = 'alert ' + alertClass + ' shadow-lg';
	toast.innerHTML = '<span>' + message + '</span>';

	if (isModalOpen) {
		// Insert at top of modal content
		const modalContent = document.getElementById('modal-content');
		toast.style.marginBottom = '1rem';
		modalContent.insertBefore(toast, modalContent.firstChild);
	} else {
		const container = document.getElementById('toast-container');
		container.appendChild(toast);
	}

	setTimeout(function() {
		toast.remove();
	}, duration);
}

// Copy to clipboard helper
function copyToClipboard(text) {
	navigator.clipboard.writeText(text).then(function() {
		showToast('Copied to clipboard!', 'success');
	}).catch(function() {
		showToast('Failed to copy', 'error');
	});
}

// Listen for HTMX events to show modal after content loads
document.body.addEventListener('htmx:afterSwap', function(evt) {
	if (evt.detail.target.id === 'modal-content') {
		showModal();
		// Check for back-url initialization
		var backUrlEl = document.querySelector('#modal-content [data-init-back-url]');
		if (backUrlEl) {
			document.getElementById('modal-content').setAttribute('data-back-url', backUrlEl.getAttribute('data-back-url'));
			backUrlEl.remove();
		} else {
			// Clear back-url for top-level modals
			document.getElementById('modal-content').removeAttribute('data-back-url');
		}
	}
});

// Close modal on successful form submission
document.body.addEventListener('htmx:afterRequest', function(evt) {
	if (evt.detail.successful && evt.detail.target.id === 'modal-content') {
		// Form was submitted successfully, modal will close via HX-Trigger
	}
});

// Handle HX-Trigger for closing modal and showing toasts
document.body.addEventListener('closeModal', function(evt) {
	closeModal();
});

document.body.addEventListener('showToast', function(evt) {
	showToast(evt.detail.message, evt.detail.type);
});

// Registrations page state
var currentCustomerID = null;
var currentProductID = null;

function loadCustomerRegistrations(customerID) {
	if (!customerID) {
		document.getElementById('registrations-container').innerHTML =
			'<div class="text-center py-12 text-base-content/60">Select a customer to view their product registrations</div>';
		document.getElementById('features-container').classList.add('hidden');
		currentCustomerID = null;
		currentProductID = null;
		return;
	}

	currentCustomerID = customerID;
	currentProductID = null;
	document.getElementById('features-container').classList.add('hidden');

	htmx.ajax('GET', '/web/licenses/' + customerID, {
		target: '#registrations-container',
		swap: 'innerHTML'
	});
}

function selectProduct(productID, licenseKey) {
	if (!currentCustomerID) return;

	currentProductID = productID;

	// Update license key display
	var keyContainer = document.getElementById('license-key-container');
	var keyDisplay = document.getElementById('license-key-display');
	if (keyDisplay && keyContainer) {
		keyDisplay.textContent = licenseKey;
		keyContainer.classList.remove('hidden');
	}

	// Load features
	document.getElementById('features-container').classList.remove('hidden');
	htmx.ajax('GET', '/web/features/' + currentCustomerID + '/' + productID, {
		target: '#features-container',
		swap: 'innerHTML'
	});

	// Highlight selected row and get product name
	var productName = '';
	document.querySelectorAll('#registrations-table tbody tr').forEach(function(row) {
		row.classList.remove('selected');
	});
	var selectedRow = document.querySelector('#registrations-table tbody tr[data-product-id="' + productID + '"]');
	if (selectedRow) {
		selectedRow.classList.add('selected');
		// Get product name from first cell
		var firstCell = selectedRow.querySelector('td');
		if (firstCell) {
			productName = firstCell.textContent;
		}
	}

	// Update product name in license key label
	var productLabel = document.getElementById('license-key-product');
	if (productLabel) {
		productLabel.textContent = productName ? ' (' + productName + ')' : '';
	}
}

function openMachinesModal(customerID, productID) {
	htmx.ajax('GET', '/web/machines/' + customerID + '/' + productID, {
		target: '#modal-content',
		swap: 'innerHTML'
	});
}

// Re-initialize after HTMX swaps
document.body.addEventListener('htmx:afterSwap', function(evt) {
	if (evt.detail.target.id === 'registrations-container') {
		// Reset product selection when registrations reload
		currentProductID = null;
		var featuresContainer = document.getElementById('features-container');
		if (featuresContainer) {
			featuresContainer.classList.add('hidden');
		}
	}
});

// Feature form: toggle Allowed Values field based on feature type
function toggleAllowedValues() {
	var select = document.getElementById('feature-type-select');
	var input = document.getElementById('allowed-values-input');
	if (!select || !input) return;

	var isValues = (select.value === 'values');
	input.disabled = !isValues;
	if (!isValues) {
		input.value = '';
	}
}

// Registration form: auto-calculate expiration dates for subscription licenses
function calcExpirationDates() {
	var isSubscription = document.getElementById('is-subscription');
	if (!isSubscription || !isSubscription.checked) return;

	var startDate = document.getElementById('start-date').value;
	var licenseTerm = parseInt(document.getElementById('license-term').value) || 0;

	if (!startDate || licenseTerm <= 0) return;

	var date = new Date(startDate);
	date.setMonth(date.getMonth() + licenseTerm);
	date.setDate(date.getDate() - 1);

	var expDate = date.toISOString().split('T')[0];
	document.getElementById('expiration-date').value = expDate;
	document.getElementById('maint-expiration-date').value = expDate;
}
