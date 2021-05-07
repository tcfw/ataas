package client

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownMethod = errors.New("unknown API method")

	errLocked = errors.New("connection locked")
)

type ResponseError struct {
	Code     ResponseCode
	Response *CryptoComResponse
}

func (re *ResponseError) Error() string {
	var desc string

	knownDesc, found := codeToDesc[re.Code]
	if found {
		desc = knownDesc
	}

	return fmt.Sprintf("code= %d: desc=%s", re.Code, desc)
}

/*
API Response Codes (see https://exchange-docs.crypto.com/spot/index.html#response-and-reason-codes)
*/
type ResponseCode int

const (
	SUCCESS ResponseCode = 0

	SYS_ERROR          ResponseCode = 10001 //Malformed request, (E.g. not using application/json for REST)
	UNAUTHORIZED       ResponseCode = 10002 //Not authenticated, or key/signature incorrect
	IP_ILLEGAL         ResponseCode = 10003 //IP address not whitelisted
	BAD_REQUEST        ResponseCode = 10004 //Missing required fields
	USER_TIER_INVALID  ResponseCode = 10005 //Disallowed based on user tier
	TOO_MANY_REQUESTS  ResponseCode = 10006 //Requests have exceeded rate limits
	INVALID_NONCE      ResponseCode = 10007 //Nonce value differs by more than 30 seconds from server
	METHOD_NOT_FOUND   ResponseCode = 10008 //Invalid method specified
	INVALID_DATE_RANGE ResponseCode = 10009 //Invalid date range

	DUPLICATE_RECORD ResponseCode = 20001 //Duplicated record
	NEGATIVE_BALANCE ResponseCode = 20002 //Insufficient balance

	SYMBOL_NOT_FOUND           ResponseCode = 30003 //Invalid instrument_name specified
	SIDE_NOT_SUPPORTED         ResponseCode = 30004 //Invalid side specified
	ORDERTYPE_NOT_SUPPORTED    ResponseCode = 30005 //Invalid type specified
	MIN_PRICE_VIOLATED         ResponseCode = 30006 //Price is lower than the minimum
	MAX_PRICE_VIOLATED         ResponseCode = 30007 //Price is higher than the maximum
	MIN_QUANTITY_VIOLATED      ResponseCode = 30008 //Quantity is lower than the minimum
	MAX_QUANTITY_VIOLATED      ResponseCode = 30009 //Quantity is higher than the maximum
	MISSING_ARGUMENT           ResponseCode = 30010 //Required argument is blank or missing
	INVALID_PRICE_PRECISION    ResponseCode = 30013 //Too many decimal places for Price
	INVALID_QUANTITY_PRECISION ResponseCode = 30014 //Too many decimal places for Quantity
	MIN_NOTIONAL_VIOLATED      ResponseCode = 30016 //The notional amount is less than the minimum
	MAX_NOTIONAL_VIOLATED      ResponseCode = 30017 //The notional amount exceeds the maximum
	MIN_AMOUNT_VIOLATED        ResponseCode = 30023 //Amount is lower than the minimum
	MAX_AMOUNT_VIOLATED        ResponseCode = 30024 //Amount is higher than the maximum
	AMOUNT_PRECISION_OVERFLOW  ResponseCode = 30025 //Amount precision exceeds the maximum

	MG_INVALID_ACCOUNT_STATUS ResponseCode = 40001 //Operation has failed due to your account's status. Please try again later.
	MG_TRANSFER_ACTIVE_LOAN   ResponseCode = 40002 //Transfer has failed due to holding an active loan. Please repay your loan and try again later.
	MG_INVALID_LOAN_CURRENCY  ResponseCode = 40003 //Currency is not same as loan currency of active loan
	MG_INVALID_REPAY_AMOUNT   ResponseCode = 40004 //Only supporting full repayment of all margin loans
	MG_NO_ACTIVE_LOAN         ResponseCode = 40005 //No active loan
	MG_BLOCKED_BORROW         ResponseCode = 40006 //Borrow has been suspended. Please try again later.
	MG_BLOCKED_NEW_ORDER      ResponseCode = 40007 //Placing new order has been suspended. Please try again later.

	DW_CREDIT_LINE_NOT_MAINTAINED ResponseCode = 50001 //Please ensure your credit line is maintained and try again later.
)

var (
	codeToDesc = map[ResponseCode]string{
		SUCCESS:                       "Success",
		SYS_ERROR:                     "Malformed request, (E.g. not using application/json for REST)",
		UNAUTHORIZED:                  "Not authenticated, or key/signature incorrect",
		IP_ILLEGAL:                    "IP address not whitelisted",
		BAD_REQUEST:                   "Missing required fields",
		USER_TIER_INVALID:             "Disallowed based on user tier",
		TOO_MANY_REQUESTS:             "Requests have exceeded rate limits",
		INVALID_NONCE:                 "Nonce value differs by more than 30 seconds from server",
		METHOD_NOT_FOUND:              "Invalid method specified",
		INVALID_DATE_RANGE:            "Invalid date range",
		DUPLICATE_RECORD:              "Duplicated record",
		NEGATIVE_BALANCE:              "Insufficient balance",
		SYMBOL_NOT_FOUND:              "Invalid instrument_name specified",
		SIDE_NOT_SUPPORTED:            "Invalid side specified",
		ORDERTYPE_NOT_SUPPORTED:       "Invalid type specified",
		MIN_PRICE_VIOLATED:            "Price is lower than the minimum",
		MAX_PRICE_VIOLATED:            "Price is higher than the maximum",
		MIN_QUANTITY_VIOLATED:         "Quantity is lower than the minimum",
		MAX_QUANTITY_VIOLATED:         "Quantity is higher than the maximum",
		MISSING_ARGUMENT:              "Required argument is blank or missing",
		INVALID_PRICE_PRECISION:       "Too many decimal places for Price",
		INVALID_QUANTITY_PRECISION:    "Too many decimal places for Quantity",
		MIN_NOTIONAL_VIOLATED:         "The notional amount is less than the minimum",
		MAX_NOTIONAL_VIOLATED:         "The notional amount exceeds the maximum",
		MIN_AMOUNT_VIOLATED:           "Amount is lower than the minimum",
		MAX_AMOUNT_VIOLATED:           "Amount is higher than the maximum",
		AMOUNT_PRECISION_OVERFLOW:     "Amount precision exceeds the maximum",
		MG_INVALID_ACCOUNT_STATUS:     "Operation has failed due to your account's status. Please try again later.",
		MG_TRANSFER_ACTIVE_LOAN:       "Transfer has failed due to holding an active loan. Please repay your loan and try again later.",
		MG_INVALID_LOAN_CURRENCY:      "Currency is not same as loan currency of active loan",
		MG_INVALID_REPAY_AMOUNT:       "Only supporting full repayment of all margin loans",
		MG_NO_ACTIVE_LOAN:             "No active loan",
		MG_BLOCKED_BORROW:             "Borrow has been suspended. Please try again later.",
		MG_BLOCKED_NEW_ORDER:          "Placing new order has been suspended. Please try again later.",
		DW_CREDIT_LINE_NOT_MAINTAINED: "Please ensure your credit line is maintained and try again later.",
	}
)
