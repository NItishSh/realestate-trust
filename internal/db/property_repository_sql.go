package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type SQLPropertyRepository struct {
	db *sql.DB
}

func NewSQLPropertyRepository(db *sql.DB) *SQLPropertyRepository {
	return &SQLPropertyRepository{db: db}
}

func (r *SQLPropertyRepository) ListProperties() ([]*Property, error) {
	query := `SELECT id, address, description, value, thumbnail, owner_id, documents,
	                 title_insurance_status, title_insurance_policy, title_insurance_company,
	                 title_insurance_verified_at, sqft, bedrooms, bathrooms, year_built,
	                 property_type, created_at
	          FROM properties`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Property
	for rows.Next() {
		p := &Property{}
		err := rows.Scan(
			&p.ID,
			&p.Address,
			&p.Description,
			&p.Value,
			&p.Thumbnail,
			&p.OwnerID,
			pq.Array(&p.Documents),
			&p.TitleInsuranceStatus,
			&p.TitleInsurancePolicy,
			&p.TitleInsuranceCompany,
			&p.TitleInsuranceVerifiedAt,
			&p.SqFt,
			&p.Bedrooms,
			&p.Bathrooms,
			&p.YearBuilt,
			&p.PropertyType,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*Property{}, nil
	}
	return result, nil
}

func (r *SQLPropertyRepository) GetProperty(id string) (*Property, error) {
	query := `SELECT id, address, description, value, thumbnail, owner_id, documents,
	                 title_insurance_status, title_insurance_policy, title_insurance_company,
	                 title_insurance_verified_at, sqft, bedrooms, bathrooms, year_built,
	                 property_type, created_at
	          FROM properties WHERE id = $1`

	p := &Property{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Address,
		&p.Description,
		&p.Value,
		&p.Thumbnail,
		&p.OwnerID,
		pq.Array(&p.Documents),
		&p.TitleInsuranceStatus,
		&p.TitleInsurancePolicy,
		&p.TitleInsuranceCompany,
		&p.TitleInsuranceVerifiedAt,
		&p.SqFt,
		&p.Bedrooms,
		&p.Bathrooms,
		&p.YearBuilt,
		&p.PropertyType,
		&p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("property not found")
		}
		return nil, err
	}
	return p, nil
}

func (r *SQLPropertyRepository) CreateProperty(address, description string, value float64, thumbnail, ownerId string) (*Property, error) {
	id := "prop-" + uuid.New().String()[:8]
	docs := []string{"Deed of Trust", "Property Inspection Report"}

	query := `INSERT INTO properties (
	              id, address, description, value, thumbnail, owner_id, documents,
	              title_insurance_status, title_insurance_policy, title_insurance_company,
	              title_insurance_verified_at, sqft, bedrooms, bathrooms, year_built,
	              property_type
	          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	          RETURNING created_at`

	p := &Property{
		ID:                       id,
		Address:                  address,
		Description:              description,
		Value:                    value,
		Thumbnail:                thumbnail,
		OwnerID:                  ownerId,
		Documents:                docs,
		TitleInsuranceStatus:     "UNINSURED",
		TitleInsuranceCompany:    "",
		TitleInsurancePolicy:     "",
		TitleInsuranceVerifiedAt: "",
	}

	err := r.db.QueryRow(query,
		id, address, description, value, thumbnail, ownerId, pq.Array(docs),
		p.TitleInsuranceStatus, p.TitleInsurancePolicy, p.TitleInsuranceCompany,
		p.TitleInsuranceVerifiedAt, p.SqFt, p.Bedrooms, p.Bathrooms, p.YearBuilt,
		p.PropertyType,
	).Scan(&p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *SQLPropertyRepository) CreatePropertyWithID(id, address, description string, value float64, thumbnail, ownerId string) (*Property, error) {
	if id == "" {
		id = "prop-" + uuid.New().String()[:8]
	}
	docs := []string{"Deed of Trust", "Property Inspection Report"}

	query := `INSERT INTO properties (
	              id, address, description, value, thumbnail, owner_id, documents,
	              title_insurance_status, title_insurance_policy, title_insurance_company,
	              title_insurance_verified_at, sqft, bedrooms, bathrooms, year_built,
	              property_type
	          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	          RETURNING created_at`

	p := &Property{
		ID:                       id,
		Address:                  address,
		Description:              description,
		Value:                    value,
		Thumbnail:                thumbnail,
		OwnerID:                  ownerId,
		Documents:                docs,
		TitleInsuranceStatus:     "UNINSURED",
		TitleInsuranceCompany:    "",
		TitleInsurancePolicy:     "",
		TitleInsuranceVerifiedAt: "",
	}

	err := r.db.QueryRow(query,
		id, address, description, value, thumbnail, ownerId, pq.Array(docs),
		p.TitleInsuranceStatus, p.TitleInsurancePolicy, p.TitleInsuranceCompany,
		p.TitleInsuranceVerifiedAt, p.SqFt, p.Bedrooms, p.Bathrooms, p.YearBuilt,
		p.PropertyType,
	).Scan(&p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *SQLPropertyRepository) UpdateTitleInsurance(id, status, company, policy string) (*Property, error) {
	query := `UPDATE properties
	          SET title_insurance_status = $1, title_insurance_company = $2,
	              title_insurance_policy = $3, title_insurance_verified_at = $4,
	              updated_at = NOW()
	          WHERE id = $5`

	verifiedAt := time.Now().Format(time.RFC3339)
	res, err := r.db.Exec(query, status, company, policy, verifiedAt, id)
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, errors.New("property not found")
	}
	return r.GetProperty(id)
}

func (r *SQLPropertyRepository) UpdatePropertyDetails(id string, sqft, bedrooms, bathrooms, yearBuilt int, propType string) (*Property, error) {
	query := `UPDATE properties
	          SET sqft = $1, bedrooms = $2, bathrooms = $3, year_built = $4,
	              property_type = $5, updated_at = NOW()
	          WHERE id = $6`

	res, err := r.db.Exec(query, sqft, bedrooms, bathrooms, yearBuilt, propType, id)
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, errors.New("property not found")
	}
	return r.GetProperty(id)
}
