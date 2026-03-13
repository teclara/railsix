export interface Stop {
	id: string;
	code: string;
	name: string;
}

export interface Alert {
	headline: string;
	description: string;
	routeNames?: string[];
}

export interface Departure {
	line: string;
	lineName?: string;
	scheduledTime: string;
	actualTime?: string;
	arrivalTime?: string;
	status: string;
	platform?: string;
	delayMinutes?: number;
	stops?: string[];
	lastStopCode?: string;
	cars?: string;
	isInMotion?: boolean;
	isCancelled?: boolean;
	isExpress?: boolean;
	alert?: string;
	routeType?: number;
	tripNumber?: string;
}

export interface DeparturesResult {
	stationAlert?: string;
	departures: Departure[];
}

export interface UnionDeparture {
	service: string;
	serviceType?: string;
	platform: string;
	time: string;
	info: string;
	stops: string[];
	cars?: string;
	isInMotion?: boolean;
	isCancelled?: boolean;
	alert?: string;
}

export interface NetworkLine {
	lineCode: string;
	lineName: string;
	activeTrips: number;
}
