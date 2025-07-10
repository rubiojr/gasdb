#!/bin/bash

# Script to regenerate CSV dumps from the fuel prices database

set -e

# Configuration
DB_PATH="db/fuel_prices.db"
DUMPS_DIR="dumps"

# Create dumps directory if it doesn't exist
mkdir -p "$DUMPS_DIR"

echo "Regenerating CSV dumps from $DB_PATH..."

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database file $DB_PATH not found!"
    exit 1
fi

# Check database integrity
# echo "Checking database integrity..."
# sqlite3 "$DB_PATH" "PRAGMA quick_check;" > /dev/null
# if [ $? -ne 0 ]; then
#     echo "Error: Database integrity check failed!"
#     exit 1
# fi

# Generate Gasolina 95 E10 CSV
echo "Generating precio_gasolina_95_e5.csv..."
sqlite3 -header -csv "$DB_PATH" "
SELECT 
    date, 
    ideess, 
    cp, 
    direccion, 
    horario, 
    latitud, 
    localidad, 
    longitud, 
    margen, 
    municipio, 
    provincia, 
    rotulo, 
    tipo_venta, 
    precio_gasolina_95_e5, 
    id_municipio, 
    id_provincia, 
    id_ccaa 
FROM historic_prices 
WHERE precio_gasolina_95_e5 IS NOT NULL 
    AND precio_gasolina_95_e5 != '' 
    AND ideess IN ('4413', '9601', '8350', '1947', '5633', '5572', '120', '33', '966', '14742', '5935', '5844', '2608', '2724', '2084', '10904', '15158', '154', '6046', '4753', '1017', '420', '3513', '1939', '11303', '10513', '14695', '4380', '619', '671', '11054', '4435', '7587', '7565', '6900', '5571', '11327', '5570', '123', '72', '12732', '1997', '6488', '5584', '13205', '4429', '6936', '6616', '15522', '1269', '13268', '8094', '4734', '11318', '3040', '2020', '8033', '251', '15062', '3111', '10903', '15406', '2106', '7456', '8114', '12995', '10645', '285', '6378', '4751', '9151', '7742', '265', '694', '2384', '9798', '6228', '5512', '7830', '7675', '6395', '10777', '14117', '11161', '6310', '8928', '12267', '10689', '1989', '11413', '5439', '5389', '3908', '3758', '6178', '11330', '9470', '6083', '1518', '9016', '4914', '5053')
ORDER BY date, ideess;
" > "$DUMPS_DIR/precio_gasolina_95_e5.csv"

# Generate Gasoleo A CSV
echo "Generating precio_gasoleo_a.csv..."
sqlite3 -header -csv "$DB_PATH" "
SELECT 
    date, 
    ideess, 
    cp, 
    direccion, 
    horario, 
    latitud, 
    localidad, 
    longitud, 
    margen, 
    municipio, 
    provincia, 
    rotulo, 
    tipo_venta, 
    precio_gasoleo_a, 
    id_municipio, 
    id_provincia, 
    id_ccaa 
FROM historic_prices 
WHERE precio_gasoleo_a IS NOT NULL 
    AND precio_gasoleo_a != '' 
    AND ideess IN ('4413', '9601', '8350', '1947', '5633', '5572', '120', '33', '966', '14742', '5935', '5844', '2608', '2724', '2084', '10904', '15158', '154', '6046', '4753', '1017', '420', '3513', '1939', '11303', '10513', '14695', '4380', '619', '671', '11054', '4435', '7587', '7565', '6900', '5571', '11327', '5570', '123', '72', '12732', '1997', '6488', '5584', '13205', '4429', '6936', '6616', '15522', '1269', '13268', '8094', '4734', '11318', '3040', '2020', '8033', '251', '15062', '3111', '10903', '15406', '2106', '7456', '8114', '12995', '10645', '285', '6378', '4751', '9151', '7742', '265', '694', '2384', '9798', '6228', '5512', '7830', '7675', '6395', '10777', '14117', '11161', '6310', '8928', '12267', '10689', '1989', '11413', '5439', '5389', '3908', '3758', '6178', '11330', '9470', '6083', '1518', '9016', '4914', '5053')
ORDER BY date, ideess;
" > "$DUMPS_DIR/precio_gasoleo_a.csv"

# Show results
echo "CSV dumps generated successfully:"
echo "- $(wc -l < "$DUMPS_DIR/precio_gasolina_95_e5.csv") lines in precio_gasolina_95_e5.csv"
echo "- $(wc -l < "$DUMPS_DIR/precio_gasoleo_a.csv") lines in precio_gasoleo_a.csv"

echo "Files saved to $DUMPS_DIR/"
