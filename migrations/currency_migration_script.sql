-- Migration script for adding currency_code and converting formatted values
DO $$
DECLARE
    prize_tiers_value_type TEXT;
    prizes_value_type TEXT;
BEGIN
    -- Get current column types
    SELECT data_type INTO prize_tiers_value_type 
    FROM information_schema.columns 
    WHERE table_name = 'prize_tiers' AND column_name = 'value';
    
    SELECT data_type INTO prizes_value_type 
    FROM information_schema.columns 
    WHERE table_name = 'prizes' AND column_name = 'value';
    
    -- If prize_tiers.value is not numeric, convert it
    IF prize_tiers_value_type NOT LIKE '%numeric%' AND prize_tiers_value_type NOT LIKE '%decimal%' THEN
        -- Add temporary column for numeric values
        ALTER TABLE prize_tiers ADD COLUMN value_numeric DECIMAL(20, 2);
        
        -- Add currency_code column
        ALTER TABLE prize_tiers ADD COLUMN currency_code VARCHAR(10) DEFAULT 'NGN';
        
        -- Create function to extract numeric value from formatted currency string
        CREATE OR REPLACE FUNCTION extract_numeric_value(value_str TEXT) 
        RETURNS DECIMAL AS $$
        DECLARE
            numeric_value DECIMAL;
        BEGIN
            -- Remove currency symbol (N), commas, and any other non-numeric characters except decimal point
            numeric_value := NULLIF(regexp_replace(value_str, '[^0-9.]', '', 'g'), '')::DECIMAL;
            RETURN numeric_value;
        EXCEPTION WHEN OTHERS THEN
            -- Return 0 if conversion fails
            RETURN 0;
        END;
        $$ LANGUAGE plpgsql;
        
        -- Update temporary column with extracted numeric values
        UPDATE prize_tiers 
        SET value_numeric = extract_numeric_value(value::TEXT)
        WHERE value IS NOT NULL;
        
        -- Drop original column and rename temporary column
        ALTER TABLE prize_tiers DROP COLUMN value;
        ALTER TABLE prize_tiers RENAME COLUMN value_numeric TO value;
        
        -- Drop the function as it's no longer needed
        DROP FUNCTION IF EXISTS extract_numeric_value(TEXT);
    ELSE
        -- Just add currency_code column if value is already numeric
        ALTER TABLE prize_tiers ADD COLUMN IF NOT EXISTS currency_code VARCHAR(10) DEFAULT 'NGN';
    END IF;
    
    -- If prizes.value is not numeric, convert it
    IF prizes_value_type NOT LIKE '%numeric%' AND prizes_value_type NOT LIKE '%decimal%' THEN
        -- Add temporary column for numeric values
        ALTER TABLE prizes ADD COLUMN value_numeric DECIMAL(20, 2);
        
        -- Add currency_code column
        ALTER TABLE prizes ADD COLUMN currency_code VARCHAR(10) DEFAULT 'NGN';
        
        -- Create function to extract numeric value from formatted currency string
        CREATE OR REPLACE FUNCTION extract_numeric_value(value_str TEXT) 
        RETURNS DECIMAL AS $$
        DECLARE
            numeric_value DECIMAL;
        BEGIN
            -- Remove currency symbol (N), commas, and any other non-numeric characters except decimal point
            numeric_value := NULLIF(regexp_replace(value_str, '[^0-9.]', '', 'g'), '')::DECIMAL;
            RETURN numeric_value;
        EXCEPTION WHEN OTHERS THEN
            -- Return 0 if conversion fails
            RETURN 0;
        END;
        $$ LANGUAGE plpgsql;
        
        -- Update temporary column with extracted numeric values
        UPDATE prizes 
        SET value_numeric = extract_numeric_value(value::TEXT)
        WHERE value IS NOT NULL;
        
        -- Drop original column and rename temporary column
        ALTER TABLE prizes DROP COLUMN value;
        ALTER TABLE prizes RENAME COLUMN value_numeric TO value;
        
        -- Drop the function as it's no longer needed
        DROP FUNCTION IF EXISTS extract_numeric_value(TEXT);
    ELSE
        -- Just add currency_code column if value is already numeric
        ALTER TABLE prizes ADD COLUMN IF NOT EXISTS currency_code VARCHAR(10) DEFAULT 'NGN';
    END IF;
END $$;
