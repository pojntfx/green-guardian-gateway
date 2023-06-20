# GreenGuardian MQTT Protocol

## Topology

A gateway has many rooms and plants. It represents a single customer, acts as a namespace and (de-)muxes messages.

- Gateway (customer)
  - Room[]
  - Plant[]

A room has one fan and one temperature sensor

- Room
  - Temperature sensor
  - Fan

A plant has one moisture sensor and one sprinkler.

- Plant
  - Moisture sensor
  - Sprinkler

## Messages

### Sensors → Gateway

**Temperature Sensor**:

```yaml
# Via TCP
roomID: 1
measurement: 24
default: 20
```

**Moisture Sensor**:

```yaml
# Via TCP
plantID: 1
measurement: 65
default: 50
```

### Actuators → Gateway

**Fan (Registration)**:

```yaml
# Via TCP. Use the `roomID` to store the connection for this room's fan in the gateway in a map.
roomID: 1
```

**Sprinkler (Registration)**:

```yaml
# Via TCP. Use the `plantID` to store the connection for this plant's sprinkler in the gateway in a map.
plantID: 1
```

### Gateway → Cloud

**Temperature Sensor**:

```yaml
# To MQTT channel: /gateways/<gatewayID>/rooms/<roomID>/temperature
measurement: 24
default: 20
```

**Moisture Sensor**:

```yaml
# To MQTT channel: /gateways/<gatewayID>/plants/<plantID>/moisture
measurement: 65
default: 50
```

### Cloud → Gateway

**Fan**:

```yaml
# To MQTT channel: /gateways/<gatewayID>/rooms/<roomID>/fan
on: true
```

**Sprinkler**:

```yaml
# To MQTT channel: /gateways/<gatewayID>/plants/<plantID>/sprinkler
on: true
```

### Gateway → Actuators

**Fan**:

```yaml
# Via TCP. Find the room's fan's connection via the map as described above.
on: true
```

**Sprinkler**:

```yaml
# Via TCP. Find the plant's sprinkler's connection via the map as described above.
on: true
```
