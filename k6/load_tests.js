import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        cep: {
            executor: 'constant-vus',
            vus: 30,
            duration: '30s',
            exec: 'testCep',
        },
        endereco: {
            executor: 'constant-vus',
            vus: 30,
            duration: '30s',
            exec: 'testEndereco',
        },
        state: {
            executor: 'constant-vus',
            vus: 30,
            duration: '30s',
            exec: 'testState',
        },
        calculateRoute: {
            executor: 'per-vu-iterations',
            iterations: 5,
             vus: 5,
             maxDuration: '30s',
            exec: 'testCalculateRoute',
        },
        calculateCep: {
            executor: 'per-vu-iterations',
            iterations: 5,
             vus: 5,
             maxDuration: '30s',
            exec: 'testCalculateCep',
        },
         calculateCoordinates: {
             executor: 'per-vu-iterations',
             iterations: 5,
             vus: 5,
             maxDuration: '30s',
             exec: 'testCalculateCoordinates',
         },
    },
    thresholds: {
        http_req_failed: ['rate<0.9'],        // <1% errors
        http_req_duration: ['p(95)<200'],      // 95% of requests <200ms
  },
}

export function login(username, password) {
    const url = 'http://3.238.87.0:7070/login';
    const payload = JSON.stringify({
        username: username,
        password: password,
    });

    const headers = {
        'Content-Type': 'application/json',
    };

    const res = http.post(url, payload, { headers });
    
    let data = res.json();

    if (res.status !== 200) {
        throw new Error(`Login failed: ${data.message}`);
    }

    return data.token;
}

export function setup() {
    const token = login("renan.gamero1313@gmail.com", "Guinho@01");

    return { token };
}

export function testCep() {
    const ceps = ["01001000", "20040002", "30140071"];
    const cep = ceps[Math.floor(Math.random() * ceps.length)];
    const url = `http://3.238.87.0:7070/address/find/${cep}`

    let res = http.get(url)

    if (res.status !== 200) {
        console.log(`Status: {${routeRes.status}}`)
        console.log(`Body: {${routeRes.body}}`)
    }

    check(res, {
        "cep status 200": (r) => r.status === 200,
    });

    sleep(1);
}

export function testEndereco() {
    const enderecos = [
        { rua: "Avenida Paulista", bairro: "Bela Vista", numero: "1000" },
        { rua: "Rua das Flores", bairro: "Jardim América", numero: "500" },
        { rua: "Avenida Brasil", bairro: "Centro", numero: "200" }
    ];
    const endereco = enderecos[Math.floor(Math.random() * enderecos.length)];
    const query = encodeURIComponent(`${endereco.rua}, ${endereco.bairro}, ${endereco.numero}`);

    const enderecoRes = http.get(
        `http://3.238.87.0:7070/address/find/v2?q=${query}`
    )

    if (enderecoRes.status !== 200) {
        console.log(`Status: {${enderecoRes.status}}`)
        console.log(`Body: {${enderecoRes.body}}`)
    }
    

    check(enderecoRes, {
        "endereco status 200": (r) => r.status === 200,
    });
    sleep(1);
}

export function testState() {
    const stateRes = http.get('http://3.238.87.0:7070/address/state')
    check(stateRes, {
        'state status 200': (r) => r.status === 200,
    })

    sleep(1);
}

export function testCalculateRoute(data) {
    const calculateRoute = JSON.stringify({
        origin: "São Paulo, SP, Brasil",
        destination: "Rio de Janeiro, RJ, Brasil",
        consumptionCity: 8,
        consumptionHwy: 9.6,
        price: 5.7,
        axles: 4,
        type: "Truck",
        waypoints: [
            "Campinas, SP, Brasil"
        ],
        typeRoute: "RÁPIDA",
        route_options: {
            include_fuel_stations: true,
            include_route_map: true,
            include_toll_costs: true,
            include_weigh_stations: true,
            include_freight_calc: true,
            include_polyline: true
        }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${data.token}`
        },
    };

    const routeRes = http.post('http://3.238.87.0:7070/check-route-tolls-easy', calculateRoute, params)

    if (routeRes.status !== 200) {
        console.log(`Status Calculate Route: {${routeRes.status}}`)
        console.log(`Body Calculate Route: {${routeRes.body}}`)
    }
    

    check(routeRes, {
        'calculate route status 200': (r) => r.status === 200,
    })

    sleep(1);
}

export function testCalculateCep(data) {
    const calculateCep = JSON.stringify({
            origin_cep: "01001-000",
            destination_cep: "90010-320",
            consumptionCity: 8,
            consumptionHwy: 9.6,
            price: 5.7,
            axles: 4,
            type: "Truck",
            waypoints: [
                "13015-000"
            ],
            typeRoute: "RÁPIDA",
            public_or_private: "private",
            favorite: false,
            route_options: {
                include_fuel_stations: true,
                include_route_map: true,
                include_toll_costs: true,
                include_weigh_stations: true,
                include_freight_calc: true,
                include_polyline: true
            },
            enterprise: false
        }
    );
    const params = {
        headers:{
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${data.token}`
        }
    };

    const cepRes = http.post('http://3.238.87.0:7070/check-route-tolls-cep', calculateCep, params)


    if (cepRes.status !== 200) {
        console.log(`Status Calculate Cep: {${cepRes.status}}`)
        console.log(`Body Calculate Cep: {${cepRes.body}}`)
    }
    

    check(cepRes, {
        'calculate cep status 200': (r) => r.status === 200,
    });

    sleep(1);
}

export function testCalculateCoordinates(data) {
    const calculateCoordinates = JSON.stringify({
        "origin_lat": "-23.451806",
        "origin_lng": "-46.712757",
        "destination_lat": "-30.035221",
        "destination_lng": "-51.226333",
        "consumptionCity": 8,
        "consumptionHwy": 9.6,
        "price": 5.7,
        "axles": 4,
        "type": "Truck",
        "waypoints": [
            {"lat": "-22.909756", "lng": "-47.05749"}
        ],
        "typeRoute": "RÁPIDA",
        "favorite": false,
        "route_options": {
            "include_fuel_stations": true,
            "include_route_map": true,
            "include_toll_costs": true,
            "include_weigh_stations": true,
            "include_freight_calc": true,
            "include_polyline": true
        }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': `Bearer ${data.token}`
        }
    };

    const coordinatesRes = http.post('http://3.238.87.0:7070/check-route-tolls-coordinate', calculateCoordinates, params)

    if (coordinatesRes.status !== 200) {
        console.log(`Status Coordinate: {${coordinatesRes.status}}`)
        console.log(`Body Coordinate: {${coordinatesRes.body}}`)
    }

    check(coordinatesRes, {
        'calculate coordinates status 200': (r) => r.status === 200,
    });

    sleep(1);
}