export interface Section {
	label: string;
	items: NavItem[];
}
export interface NavItem {
	id: string;
	label: string;
	icon: string;
	href: string;
}
