# frozen_string_literal: true

class User < ApplicationRecord
  before_save :assign_implementer_id
  before_create :check_impl

  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :trackable and :omniauthable
  devise :database_authenticatable, :registerable,
         :recoverable, :validatable, :trackable, 
         :timeoutable, :confirmable, :password_expirable,
         :password_archivable, :invitable, :expirable

  validates :first_name, :last_name, :implementer, presence: true
  validates :email, domain_exists: true
  validate :password_complexity
  validates :agree_to_terms, inclusion: {
    in: [true], message: 'you must agree to the terms of service to create an account'
  }

  def check_impl
    @host = self.invited_by_id

    return if @host.nil?

    @user = self
    @invite = User.where(id: @host).first
    user_id = @user.implementer_id
    invite_id = @invite.implementer_id
    user_imp = @user.implementer
    invite_imp = @invite.implementer

    if user_id != invite_id
      @user.implementer_id = invite_id
    end

    if user_imp != invite_imp
      @user.implementer = invite_imp
    end
  end

  def name
    "#{first_name} #{last_name}"
  end

  private

  # TODO: remove after connecting to API
  def assign_implementer_id
    self.implementer_id = SecureRandom.uuid if implementer_id.blank?
  end

  def password_complexity
    return if password.nil?

    password_regex = /(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*[!@\#\$\&*\-])/

    return if password.match? password_regex

    errors.add :password, 'must include at least one number, one lowercase letter,
                           one uppercase letter, and one special character (!@#$&*-)'
  end
end
